package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"sing-panel/models"

	"github.com/google/uuid"
)

const syncHeader = "X-Sync-Token"

// InstanceService manages remote panel instances and configuration sync.
//
// Sync is built on top of the existing database export/import feature: the
// config fingerprint is computed from the exported data (with machine-local
// buckets/keys excluded), and pushing/pulling transfers the export through
// the remote panel's /api/db/import and /api/db/export endpoints.
type InstanceService struct {
	db             *Database
	configService  *ConfigService
	processService *ProcessService
	kernelService  *KernelService
	start          time.Time
}

func NewInstanceService(db *Database, configService *ConfigService, processService *ProcessService, kernelService *KernelService) *InstanceService {
	return &InstanceService{
		db:             db,
		configService:  configService,
		processService: processService,
		kernelService:  kernelService,
		start:          time.Now(),
	}
}

// ---- CRUD ----

func (s *InstanceService) List() ([]models.ManagedInstance, error) {
	export, err := s.db.ExportAll()
	if err != nil {
		return nil, err
	}
	instances := make([]models.ManagedInstance, 0, len(export["instances"]))
	for _, raw := range export["instances"] {
		var inst models.ManagedInstance
		if err := json.Unmarshal([]byte(raw), &inst); err != nil {
			slog.Error("skip invalid instance record", "error", err)
			continue
		}
		instances = append(instances, inst)
	}
	sort.Slice(instances, func(i, j int) bool { return instances[i].Name < instances[j].Name })
	return instances, nil
}

func (s *InstanceService) Get(ID string) (models.ManagedInstance, error) {
	var inst models.ManagedInstance
	if err := s.db.Get("instances", ID, &inst); err != nil {
		return models.ManagedInstance{}, fmt.Errorf("instance not found")
	}
	return inst, nil
}

// normalizeInstanceURL trims and prepends http:// when no protocol is given.
func normalizeInstanceURL(raw string) string {
	url := strings.TrimRight(strings.TrimSpace(raw), "/")
	if url == "" {
		return ""
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	return url
}

func (s *InstanceService) Create(inst models.ManagedInstance) (models.ManagedInstance, error) {
	// Name is optional: when empty it is filled with the remote panel's
	// hostname on the first successful check.
	inst.URL = normalizeInstanceURL(inst.URL)
	if inst.URL == "" {
		return models.ManagedInstance{}, fmt.Errorf("地址不能为空")
	}
	if inst.Timeout <= 0 {
		inst.Timeout = 10
	}
	now := time.Now()
	inst.ID = uuid.New().String()
	inst.CreatedAt = now
	inst.UpdatedAt = now
	if err := s.db.Put("instances", inst.ID, inst); err != nil {
		return models.ManagedInstance{}, err
	}
	return inst, nil
}

func (s *InstanceService) Update(ID string, update models.InstanceUpdate) (models.ManagedInstance, error) {
	existing, err := s.Get(ID)
	if err != nil {
		return models.ManagedInstance{}, err
	}
	if strings.TrimSpace(update.Name) != "" {
		existing.Name = update.Name
	}
	if strings.TrimSpace(update.URL) != "" {
		existing.URL = normalizeInstanceURL(update.URL)
	}
	if update.Token != nil {
		existing.Token = *update.Token
	}
	if update.Timeout > 0 {
		existing.Timeout = update.Timeout
	}
	existing.UpdatedAt = time.Now()
	if err := s.db.Put("instances", existing.ID, existing); err != nil {
		return models.ManagedInstance{}, err
	}
	return existing, nil
}

func (s *InstanceService) Delete(ID string) error {
	if !s.db.HasKey("instances", ID) {
		return fmt.Errorf("instance not found")
	}
	return s.db.Delete("instances", ID)
}

// ---- Sync token ----
// The sync token is an instance-local secret. It is stored in the `state`
// bucket, which ImportAll always preserves, so syncing configurations never
// overwrites it and it never leaves this panel.

const syncTokenKey = "sync_token"

func (s *InstanceService) SyncToken() string {
	var token string
	if err := s.db.Get("state", syncTokenKey, &token); err != nil {
		return ""
	}
	return token
}

func (s *InstanceService) SetSyncToken(token string) error {
	if err := s.db.Put("state", syncTokenKey, token); err != nil {
		return err
	}
	slog.Info("sync token updated")
	return nil
}

// ---- Fingerprint ----

// configBuckets and configCacheKeys define what counts as "configuration".
// Everything else (runtime state, managed instances, tree caches) is
// machine-local and excluded from the fingerprint.
var configBuckets = map[string]bool{
	"config":  true,
	"singbox": true,
}

var singboxCacheKeys = map[string]bool{
	"geo_tree_cache":    true,
	"common_tree_cache": true,
}

// ComputeFingerprint produces a deterministic hash over the exported
// configuration data so that the same configuration yields the same value on
// every panel, regardless of machine-local state.
func ComputeFingerprint(export map[string]map[string]string) string {
	bucketNames := make([]string, 0, len(export))
	for name := range export {
		if !configBuckets[name] {
			continue
		}
		bucketNames = append(bucketNames, name)
	}
	sort.Strings(bucketNames)

	h := sha256.New()
	for _, name := range bucketNames {
		keys := export[name]
		keyNames := make([]string, 0, len(keys))
		for k := range keys {
			if name == "singbox" && singboxCacheKeys[k] {
				continue
			}
			keyNames = append(keyNames, k)
		}
		sort.Strings(keyNames)
		for _, k := range keyNames {
			value := keys[k]
			if name == "config" && k == appConfigKey {
				value = stripAppConfigDashboards(value)
			}
			h.Write([]byte(name))
			h.Write([]byte{0})
			h.Write([]byte(k))
			h.Write([]byte{0})
			h.Write([]byte(value))
			h.Write([]byte{0})
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

// LocalFingerprint returns the fingerprint of this panel's configuration.
func (s *InstanceService) LocalFingerprint() (string, error) {
	export, err := s.db.ExportAll()
	if err != nil {
		return "", err
	}
	return ComputeFingerprint(export), nil
}

// LocalExport returns the syncable export: full database minus
// machine-local buckets/keys that must not overwrite the remote panel's
// state.
func (s *InstanceService) LocalExport() (map[string]map[string]string, error) {
	export, err := s.db.ExportAll()
	if err != nil {
		return nil, err
	}
	return sanitizeExport(export), nil
}

// sanitizeExport returns the syncable export: machine-local buckets that
// must not overwrite the remote panel's state are removed. It is applied
// both when pushing (local -> remote) and when pulling (remote -> local).
// The dashboards setting is also stripped so it is never synced.
func sanitizeExport(export map[string]map[string]string) map[string]map[string]string {
	result := make(map[string]map[string]string, len(export))
	for name, keys := range export {
		if preservedBuckets[name] {
			continue
		}
		newKeys := make(map[string]string, len(keys))
		for k, v := range keys {
			if name == "config" && k == appConfigKey {
				v = stripAppConfigDashboards(v)
			}
			newKeys[k] = v
		}
		result[name] = newKeys
	}
	return result
}

// ---- Remote HTTP helpers ----

func (s *InstanceService) httpClient(inst models.ManagedInstance) *http.Client {
	timeout := inst.Timeout
	if timeout <= 0 {
		timeout = 10
	}
	return &http.Client{Timeout: time.Duration(timeout) * time.Second}
}

func (s *InstanceService) newRequest(inst models.ManagedInstance, method, path string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, inst.URL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "sing-panel")
	if inst.Token != "" {
		req.Header.Set(syncHeader, inst.Token)
	}
	return req, nil
}

func decodeResponse(resp *http.Response) (map[string]interface{}, error) {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("remote returned invalid JSON: %v", err)
	}
	if success, _ := result["success"].(bool); !success {
		if msg, _ := result["error"].(string); msg != "" {
			return nil, fmt.Errorf("remote error: %s", msg)
		}
		return nil, fmt.Errorf("remote returned failure (http %d)", resp.StatusCode)
	}
	return result, nil
}

// fetchRemoteExport downloads the remote panel's database export.
func (s *InstanceService) fetchRemoteExport(inst models.ManagedInstance) (map[string]map[string]string, error) {
	req, err := s.newRequest(inst, http.MethodGet, "/api/db/export", nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient(inst).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, fmt.Errorf("认证失败（同步令牌不匹配）")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("remote returned http %d", resp.StatusCode)
	}
	result, err := decodeResponse(resp)
	if err != nil {
		return nil, err
	}
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("remote export data missing")
	}
	export := make(map[string]map[string]string, len(data))
	for name, bucketVal := range data {
		bucket, ok := bucketVal.(map[string]interface{})
		if !ok {
			continue
		}
		keys := make(map[string]string, len(bucket))
		for k, v := range bucket {
			if str, ok := v.(string); ok {
				keys[k] = str
			} else if raw, err := json.Marshal(v); err == nil {
				keys[k] = string(raw)
			}
		}
		export[name] = keys
	}
	return export, nil
}

// fetchRemoteInfo downloads the remote panel's basic information.
func (s *InstanceService) fetchRemoteInfo(inst models.ManagedInstance) (*models.PanelInfo, error) {
	req, err := s.newRequest(inst, http.MethodGet, "/api/panel/info", nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient(inst).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, fmt.Errorf("认证失败（同步令牌不匹配）")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("remote returned http %d", resp.StatusCode)
	}
	result, err := decodeResponse(resp)
	if err != nil {
		return nil, err
	}
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("remote info data missing")
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var info models.PanelInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, fmt.Errorf("invalid remote info: %v", err)
	}
	return &info, nil
}

// ---- Status check ----

// CheckInstance fetches the remote panel's info and export in parallel and
// compares its config fingerprint with the local one.
func (s *InstanceService) CheckInstance(inst models.ManagedInstance) models.InstanceStatus {
	status := models.InstanceStatus{
		Instance:  inst,
		CheckedAt: time.Now(),
	}
	start := time.Now()

	localFingerprint, err := s.LocalFingerprint()
	if err != nil {
		status.Error = fmt.Sprintf("本地指纹计算失败: %v", err)
		return status
	}

	type exportResult struct {
		export map[string]map[string]string
		err    error
	}
	exportCh := make(chan exportResult, 1)
	infoCh := make(chan *models.PanelInfo, 1)
	go func() {
		export, err := s.fetchRemoteExport(inst)
		exportCh <- exportResult{export: export, err: err}
	}()
	go func() {
		info, err := s.fetchRemoteInfo(inst)
		if err != nil {
			infoCh <- nil
			return
		}
		infoCh <- info
	}()

	res := <-exportCh
	info := <-infoCh
	if res.err != nil || info == nil {
		status.Error = res.err.Error()
		if status.Error == "" {
			status.Error = "获取面板信息失败"
		}
		return status
	}

	status.Reachable = true
	status.LatencyMs = time.Since(start).Milliseconds()
	status.Info = info
	status.Fingerprint = ComputeFingerprint(res.export)

	// Fill in the display name from the remote hostname when it was left
	// empty at creation time.
	if strings.TrimSpace(inst.Name) == "" && info.Hostname != "" {
		inst.Name = info.Hostname
		if err := s.db.Put("instances", inst.ID, inst); err == nil {
			status.Instance = inst
		}
	}

	inSync := status.Fingerprint == localFingerprint
	status.InSync = &inSync
	return status
}

// CheckAll checks all managed instances.
func (s *InstanceService) CheckAll() []models.InstanceStatus {
	instances, err := s.List()
	if err != nil {
		return nil
	}
	statuses := make([]models.InstanceStatus, 0, len(instances))
	for _, inst := range instances {
		statuses = append(statuses, s.CheckInstance(inst))
	}
	return statuses
}

// ---- Sync ----

// SyncPush exports the local configuration and imports it into the remote
// panel. The remote panel's ImportAll preserves its own local state and
// managed instances.
func (s *InstanceService) SyncPush(inst models.ManagedInstance) error {
	export, err := s.LocalExport()
	if err != nil {
		return fmt.Errorf("本地导出失败: %w", err)
	}
	payload, err := json.Marshal(map[string]interface{}{"data": export})
	if err != nil {
		return err
	}
	req, err := s.newRequest(inst, http.MethodPost, "/api/db/import", payload)
	if err != nil {
		return err
	}
	resp, err := s.httpClient(inst).Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return fmt.Errorf("认证失败（同步令牌不匹配）")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("remote returned http %d", resp.StatusCode)
	}
	result, err := decodeResponse(resp)
	if err != nil {
		return err
	}
	slog.Info("config pushed to remote instance", "name", inst.Name, "url", inst.URL)
	if msg, _ := result["message"].(string); msg != "" {
		slog.Debug("remote push message", "message", msg)
	}
	return nil
}

// SyncPushAll pushes the local configuration to every managed instance.
func (s *InstanceService) SyncPushAll() []models.InstanceStatus {
	instances, err := s.List()
	if err != nil {
		return nil
	}
	statuses := make([]models.InstanceStatus, 0, len(instances))
	for _, inst := range instances {
		status := models.InstanceStatus{Instance: inst, CheckedAt: time.Now()}
		start := time.Now()
		if err := s.SyncPush(inst); err != nil {
			status.Error = err.Error()
		} else {
			status.Reachable = true
			status.LatencyMs = time.Since(start).Milliseconds()
		}
		statuses = append(statuses, status)
	}
	return statuses
}

// SyncPull fetches the remote panel's configuration and replaces the local
// configuration with it. Local runtime state and managed instances are kept.
func (s *InstanceService) SyncPull(inst models.ManagedInstance) error {
	export, err := s.fetchRemoteExport(inst)
	if err != nil {
		return fmt.Errorf("获取远端配置失败: %w", err)
	}
	export = sanitizeExport(export)
	if err := s.db.ImportAll(export); err != nil {
		return fmt.Errorf("导入配置失败: %w", err)
	}
	s.configService.Reload()
	slog.Info("config pulled from remote instance", "name", inst.Name, "url", inst.URL)
	return nil
}

// LocalPanelInfo returns this panel's basic information.
func (s *InstanceService) LocalPanelInfo() models.PanelInfo {
	info := models.PanelInfo{
		Version:          singBoxVersion(),
		SingboxRunning:   GetBoxService().IsRunning(),
		SyncTokenEnabled: s.SyncToken() != "",
	}
	systemInfo := s.kernelService.GetSystemInfo()
	info.Hostname = systemInfo.Hostname
	info.Platform = systemInfo.Platform
	info.Arch = systemInfo.Arch
	info.KernelVersion = systemInfo.KernelVersion
	if status := s.processService.GetStatus(); status.Running {
		info.SingboxRunning = true
	}
	info.UptimeSeconds = int64(time.Since(s.start).Seconds())
	info.DBSize = s.db.DBFileSize()
	return info
}

// ---- Config Diff ----

// ConfigDiffItem represents a single configuration difference.
type ConfigDiffItem struct {
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	LocalValue  string `json:"localValue"`
	RemoteValue string `json:"remoteValue"`
	Type        string `json:"type"` // "added", "removed", "modified"
}

// ConfigDiffResult represents the full diff between local and remote configs.
type ConfigDiffResult struct {
	Differences       []ConfigDiffItem `json:"differences"`
	LocalFingerprint  string           `json:"localFingerprint"`
	RemoteFingerprint string           `json:"remoteFingerprint"`
}

// ComputeConfigDiff compares local and remote exports and returns the differences.
func ComputeConfigDiff(localExport, remoteExport map[string]map[string]string) ConfigDiffResult {
	result := ConfigDiffResult{
		LocalFingerprint:  ComputeFingerprint(localExport),
		RemoteFingerprint: ComputeFingerprint(remoteExport),
		Differences:       make([]ConfigDiffItem, 0),
	}

	// Collect all bucket names
	allBuckets := make(map[string]bool)
	for name := range localExport {
		if configBuckets[name] {
			allBuckets[name] = true
		}
	}
	for name := range remoteExport {
		if configBuckets[name] {
			allBuckets[name] = true
		}
	}

	// Compare each bucket
	for bucket := range allBuckets {
		localKeys := localExport[bucket]
		remoteKeys := remoteExport[bucket]
		if localKeys == nil {
			localKeys = make(map[string]string)
		}
		if remoteKeys == nil {
			remoteKeys = make(map[string]string)
		}

		// Collect all keys in this bucket
		allKeys := make(map[string]bool)
		for k := range localKeys {
			if bucket == "singbox" && singboxCacheKeys[k] {
				continue
			}
			allKeys[k] = true
		}
		for k := range remoteKeys {
			if bucket == "singbox" && singboxCacheKeys[k] {
				continue
			}
			allKeys[k] = true
		}

		// Compare each key
		for key := range allKeys {
			localVal := localKeys[key]
			remoteVal := remoteKeys[key]

			// Strip dashboards for comparison
			if bucket == "config" && key == appConfigKey {
				localVal = stripAppConfigDashboards(localVal)
				remoteVal = stripAppConfigDashboards(remoteVal)
			}

			if localVal == "" && remoteVal != "" {
				result.Differences = append(result.Differences, ConfigDiffItem{
					Bucket:      bucket,
					Key:         key,
					LocalValue:  "",
					RemoteValue: remoteVal,
					Type:        "added",
				})
			} else if localVal != "" && remoteVal == "" {
				result.Differences = append(result.Differences, ConfigDiffItem{
					Bucket:      bucket,
					Key:         key,
					LocalValue:  localVal,
					RemoteValue: "",
					Type:        "removed",
				})
			} else if localVal != remoteVal {
				result.Differences = append(result.Differences, ConfigDiffItem{
					Bucket:      bucket,
					Key:         key,
					LocalValue:  localVal,
					RemoteValue: remoteVal,
					Type:        "modified",
				})
			}
		}
	}

	return result
}

// GetConfigDiff fetches the remote export and computes the diff with local config.
func (s *InstanceService) GetConfigDiff(inst models.ManagedInstance) (ConfigDiffResult, error) {
	// Get local export
	localExport, err := s.db.ExportAll()
	if err != nil {
		return ConfigDiffResult{}, fmt.Errorf("本地导出失败: %w", err)
	}

	// Get remote export
	remoteExport, err := s.fetchRemoteExport(inst)
	if err != nil {
		return ConfigDiffResult{}, fmt.Errorf("获取远端配置失败: %w", err)
	}

	return ComputeConfigDiff(localExport, remoteExport), nil
}
