//go:build linux

package services

import (
	"errors"
	"log/slog"
	"net"
	"net/netip"
	"strings"

	"github.com/sagernet/netlink"
	"github.com/sagernet/nftables"
	"github.com/sagernet/nftables/binaryutil"
	"github.com/sagernet/nftables/expr"

	"golang.org/x/sys/unix"

	"sing-panel/models"
)

const (
	tproxyNftTableName = "sing_panel"
	tproxyNftChainName = "prerouting"
	tproxyRouteTable   = 100
	tproxyRulePriority = 100
	tproxyMark         = 1
)

const (
	protoTCP byte = 6
	protoUDP byte = 17
)

// tproxyPortRule describes which protocols/families a tproxy inbound listens on.
type tproxyPortRule struct {
	port int
	tcp  bool
	udp  bool
}

// setupTproxyFirewall configures transparent-proxy traffic steering for the
// enabled tproxy inbounds using the nftables and netlink syscall APIs, so no
// external iptables/nft/ip binaries are required.
func (s *ProcessService) setupTproxyFirewall(tproxyInbounds []models.Inbound) {
	if len(tproxyInbounds) == 0 {
		return
	}
	slog.Info("setting up tproxy firewall (nftables)", "count", len(tproxyInbounds))

	if err := setupTproxyNftables(tproxyInbounds); err != nil {
		slog.Error("tproxy nftables setup failed", "error", err)
	}
}

// cleanupTproxyFirewall removes nftables/netlink tproxy rules.
func (s *ProcessService) cleanupTproxyFirewall() {
	cleanupTproxyNftables()
}

func nftablesTableExists() bool {
	conn, err := nftables.New()
	if err != nil {
		return false
	}
	for _, fam := range []nftables.TableFamily{nftables.TableFamilyIPv4, nftables.TableFamilyIPv6} {
		tb, err := conn.ListTableOfFamily(tproxyNftTableName, fam)
		if err == nil && tb != nil {
			return true
		}
	}
	return false
}

// setupTproxyNftables installs nftables prerouting tables (ip + ip6) that
// redirect matching traffic to the tproxy ports, plus the ip rule/route for
// the TPROXY mark.
func setupTproxyNftables(inbounds []models.Inbound) error {
	var v4Rules, v6Rules []tproxyPortRule
	for _, inbound := range inbounds {
		if inbound.Type != "tproxy" || !inbound.Enabled {
			continue
		}
		if inbound.ListenPort <= 0 || inbound.ListenPort > 65535 {
			slog.Warn("tproxy inbound has invalid port, skipped", "tag", inbound.Tag, "port", inbound.ListenPort)
			continue
		}
		rule := tproxyPortRule{port: inbound.ListenPort}
		switch strings.ToLower(inboundNetwork(inbound)) {
		case "tcp":
			rule.tcp = true
		case "udp":
			rule.udp = true
		default:
			rule.tcp, rule.udp = true, true
		}
		if inbound.Listen == "" {
			v4Rules = append(v4Rules, rule)
			v6Rules = append(v6Rules, rule)
			continue
		}
		addr, err := netip.ParseAddr(inbound.Listen)
		if err == nil {
			if addr.Is4() {
				v4Rules = append(v4Rules, rule)
			} else {
				v6Rules = append(v6Rules, rule)
			}
		}
	}
	if len(v4Rules) == 0 && len(v6Rules) == 0 {
		return errors.New("no valid tproxy inbound to configure")
	}

	conn, err := nftables.New()
	if err != nil {
		return err
	}

	if len(v4Rules) > 0 {
		addTproxyTable(conn, nftables.TableFamilyIPv4, v4Rules)
	}
	if len(v6Rules) > 0 {
		addTproxyTable(conn, nftables.TableFamilyIPv6, v6Rules)
	}
	if err := conn.Flush(); err != nil {
		return err
	}

	if err := addTproxyRouteRules(); err != nil {
		slog.Warn("tproxy route rule (netlink) setup failed", "error", err)
	}
	return nil
}

func inboundNetwork(inbound models.Inbound) string {
	if v, ok := inbound.Options["network"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func addTproxyTable(conn *nftables.Conn, fam nftables.TableFamily, portRules []tproxyPortRule) {
	table := &nftables.Table{Family: fam, Name: tproxyNftTableName}
	conn.AddTable(table)
	chain := conn.AddChain(&nftables.Chain{
		Table:    table,
		Name:     tproxyNftChainName,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityMangle,
	})
	accept := nftables.ChainPolicyAccept
	chain.Policy = &accept

	for _, r := range tproxyPreroutingExprs(fam, portRules) {
		conn.AddRule(&nftables.Rule{Table: table, Chain: chain, Exprs: r})
	}
}

// tproxyPreroutingExprs builds the ordered prerouting rules for one family.
// Rule order matters: exclusions first, then the TPROXY redirects.
func tproxyPreroutingExprs(fam nftables.TableFamily, portRules []tproxyPortRule) [][]expr.Any {
	rules := [][]expr.Any{
		returnFibTypeExprs(fam, unix.RTN_LOCAL),
		returnMarkedExprs(),
		returnFibTypeExprs(fam, unix.RTN_BROADCAST),
		returnFibTypeExprs(fam, unix.RTN_MULTICAST),
		returnFibTypeExprs(fam, unix.RTN_ANYCAST),
		returnMulticastExprs(fam),
	}

	for _, pr := range portRules {
		if pr.tcp {
			rules = append(rules, tproxyProtoExprs(fam, protoTCP, pr.port))
		}
		if pr.udp {
			rules = append(rules, tproxyProtoExprs(fam, protoUDP, pr.port))
		}
	}
	return rules
}

// returnFibTypeExprs exits when the destination address resolves to a given
// route type (local, broadcast, multicast, anycast). This matches iptables'
// "-m addrtype --dst-type LOCAL/BROADCAST/MULTICAST" and protects the panel's
// own DNS responses, LAN broadcasts and link-local flows from interception.
func returnFibTypeExprs(fam nftables.TableFamily, addrType uint32) []expr.Any {
	return []expr.Any{
		&expr.Fib{Register: 1, ResultADDRTYPE: true, FlagDADDR: true},
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: binaryutil.NativeEndian.PutUint32(addrType)},
		&expr.Verdict{Kind: expr.VerdictReturn},
	}
}

// returnMarkedExprs stops re-interception of already marked TPROXY packets
// (prevents loops through the local routing table).
func returnMarkedExprs() []expr.Any {
	return []expr.Any{
		&expr.Meta{Key: expr.MetaKeyMARK, Register: 1},
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: binaryutil.NativeEndian.PutUint32(tproxyMark)},
		&expr.Verdict{Kind: expr.VerdictReturn},
	}
}

// returnMulticastExprs skips traffic to well-known reserved ranges that are
// not fully covered by the route-type checks above: IPv4 multicast &
// limited broadcast and the 240/4 reservation, IPv6 multicast.
func returnMulticastExprs(fam nftables.TableFamily) []expr.Any {
	if fam == nftables.TableFamilyIPv4 {
		return []expr.Any{
			// ip daddr 255.255.255.255 (limited broadcast)
			&expr.Payload{OperationType: expr.PayloadLoad, DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4},
			&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{255, 255, 255, 255}},
			&expr.Verdict{Kind: expr.VerdictReturn},
			// ip daddr 240.0.0.0/4
			&expr.Payload{OperationType: expr.PayloadLoad, DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4},
			&expr.Bitwise{SourceRegister: 1, DestRegister: 1, Len: 4, Mask: []byte{0xf0, 0, 0, 0}, Xor: []byte{0, 0, 0, 0}},
			&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{0xf0, 0, 0, 0}},
			&expr.Verdict{Kind: expr.VerdictReturn},
		}
	}
	return []expr.Any{
		// ip6 daddr ff00::/8
		&expr.Payload{
			OperationType: expr.PayloadLoad,
			DestRegister:  1,
			Base:          expr.PayloadBaseNetworkHeader,
			Offset:        24,
			Len:           16,
		},
		&expr.Bitwise{
			SourceRegister: 1, DestRegister: 1, Len: 16,
			Mask: append([]byte{0xff}, make([]byte, 15)...),
			Xor:  make([]byte, 16),
		},
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: append([]byte{0xff}, make([]byte, 15)...)},
		&expr.Verdict{Kind: expr.VerdictReturn},
	}
}

// tproxyProtoExprs builds a single protocol TPROXY rule: match protocol, set
// the tproxy mark, then redirect the packet to the local tproxy port.
func tproxyProtoExprs(fam nftables.TableFamily, proto byte, port int) []expr.Any {
	var tproxyFamily byte
	if fam == nftables.TableFamilyIPv4 {
		tproxyFamily = unix.NFPROTO_IPV4
	} else {
		tproxyFamily = unix.NFPROTO_IPV6
	}
	portReg := uint32(1)
	markReg := uint32(2)
	return []expr.Any{
		// meta l4proto <proto>
		&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{proto}},
		// meta mark set 0x1
		&expr.Immediate{Register: markReg, Data: binaryutil.NativeEndian.PutUint32(tproxyMark)},
		&expr.Meta{Key: expr.MetaKeyMARK, SourceRegister: true, Register: markReg},
		// reg port = listen port
		&expr.Immediate{Register: portReg, Data: binaryutil.BigEndian.PutUint16(uint16(port))},
		// tproxy <family> to :port
		&expr.TProxy{Family: tproxyFamily, RegAddr: 0, RegPort: portReg},
	}
}

// addTproxyRouteRules installs "fwmark 1 lookup 100" rules and the
// "local 0.0.0.0/0 dev lo table 100" routes via netlink.
func addTproxyRouteRules() error {
	loIdx, err := loopbackIndex()
	if err != nil {
		return err
	}

	for _, family := range []int{netlink.FAMILY_V4, netlink.FAMILY_V6} {
		rule := netlink.NewRule()
		rule.Family = family
		rule.Mark = tproxyMark
		rule.MarkSet = true
		rule.Mask = 0xffffffff
		rule.Table = tproxyRouteTable
		rule.Priority = tproxyRulePriority
		if err := netlink.RuleAdd(rule); err != nil && !isExist(err) {
			return err
		}

		var dst *net.IPNet
		if family == netlink.FAMILY_V4 {
			dst = &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}
		} else {
			dst = &net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, 128)}
		}
		route := &netlink.Route{
			LinkIndex: loIdx,
			Table:     tproxyRouteTable,
			Type:      unix.RTN_LOCAL,
			Scope:     netlink.SCOPE_HOST,
			Dst:       dst,
			Family:    family,
		}
		if err := netlink.RouteAdd(route); err != nil && !isExist(err) {
			return err
		}
	}
	return nil
}

func loopbackIndex() (int, error) {
	lo, err := netlink.LinkByName("lo")
	if err != nil {
		return 0, err
	}
	return lo.Attrs().Index, nil
}

// cleanupTproxyNftables deletes the nftables tables and the netlink rules.
// Reports whether the nftables table was removed cleanly.
func cleanupTproxyNftables() bool {
	cleaned := false
	conn, err := nftables.New()
	if err == nil {
		deleted := false
		for _, fam := range []nftables.TableFamily{nftables.TableFamilyIPv4, nftables.TableFamilyIPv6} {
			if tb, err := conn.ListTableOfFamily(tproxyNftTableName, fam); err == nil && tb != nil {
				conn.DelTable(tb)
				deleted = true
			}
		}
		if deleted {
			if err := conn.Flush(); err == nil {
				cleaned = true
			}
		}
	}

	// Remove netlink rules regardless
	removeTproxyRouteRules()

	return cleaned
}

func removeTproxyRouteRules() {
	for _, family := range []int{netlink.FAMILY_V4, netlink.FAMILY_V6} {
		rules, err := netlink.RuleList(family)
		if err == nil {
			for _, r := range rules {
				if r.Mark == tproxyMark && r.Table == tproxyRouteTable {
					_ = netlink.RuleDel(&r)
				}
			}
		}
		routes, err := netlink.RouteListFiltered(family, &netlink.Route{Table: tproxyRouteTable, Type: unix.RTN_LOCAL}, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_TYPE)
		if err == nil {
			for _, r := range routes {
				_ = netlink.RouteDel(&r)
			}
		}
	}
}

func isExist(err error) bool {
	return errors.Is(err, unix.EEXIST)
}
