package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	boxlog "github.com/sagernet/sing-box/log"
)

const (
	MemoryLogCapacity = 2048
	MemoryLogMaxBytes = 4 * 1024
)

type MemoryLogEntry struct {
	Seq     uint64    `json:"seq"`
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Source  string    `json:"source"`
	Message string    `json:"message"`
}

type memoryLogSlot struct {
	seq     uint64
	time    time.Time
	level   string
	source  string
	length  int
	message [MemoryLogMaxBytes]byte
}

// MemoryLogRing is a fixed-size, process-local diagnostic log buffer. Once it
// is full, new entries overwrite the oldest entries without allocating more
// storage for messages.
type MemoryLogRing struct {
	mu       sync.RWMutex
	slots    [MemoryLogCapacity]memoryLogSlot
	next     int
	count    int
	dropped  uint64
	sequence atomic.Uint64
	clients  map[*memoryLogSubscriber]struct{}
}

type memoryLogSubscriber struct {
	channel chan MemoryLogEntry
	level   string
	source  string
}

func NewMemoryLogRing() *MemoryLogRing {
	return &MemoryLogRing{clients: make(map[*memoryLogSubscriber]struct{})}
}

func (r *MemoryLogRing) Append(level, source, message string) {
	message = strings.TrimRight(message, "\r\n")
	if len(message) > MemoryLogMaxBytes {
		message = message[:MemoryLogMaxBytes]
	}
	entry := MemoryLogEntry{
		Seq:     r.sequence.Add(1),
		Time:    time.Now(),
		Level:   level,
		Source:  source,
		Message: message,
	}

	r.mu.Lock()
	slot := &r.slots[r.next]
	if r.count == MemoryLogCapacity {
		r.dropped++
	}
	slot.seq = entry.Seq
	slot.time = entry.Time
	slot.level = entry.Level
	slot.source = entry.Source
	slot.length = copy(slot.message[:], message)
	r.next = (r.next + 1) % MemoryLogCapacity
	if r.count < MemoryLogCapacity {
		r.count++
	}
	clients := make([]*memoryLogSubscriber, 0, len(r.clients))
	for client := range r.clients {
		clients = append(clients, client)
	}
	// Keep the lock while delivering. Unsubscribe also takes this lock before
	// closing a client, so this prevents a concurrent send-on-closed-channel.
	for _, client := range clients {
		if !logLevelMatches(entry.Level, client.level) || client.source != "" && client.source != entry.Source {
			continue
		}
		select {
		case client.channel <- entry:
		default:
		}
	}
	r.mu.Unlock()
}

func (r *MemoryLogRing) Recent(limit int, level, source string) []MemoryLogEntry {
	if limit <= 0 || limit > MemoryLogCapacity {
		limit = MemoryLogCapacity
	}
	level = strings.ToLower(strings.TrimSpace(level))
	source = strings.ToLower(strings.TrimSpace(source))

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.recentLocked(0, limit, level, source)
}

// RecentAfter returns at most limit entries with a sequence greater than
// afterSeq. It is used by the UI to resume a paused stream without fetching
// the whole ring again.
func (r *MemoryLogRing) RecentAfter(afterSeq uint64, limit int, level, source string) []MemoryLogEntry {
	if limit <= 0 || limit > MemoryLogCapacity {
		limit = MemoryLogCapacity
	}
	level = strings.ToLower(strings.TrimSpace(level))
	source = strings.ToLower(strings.TrimSpace(source))

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.recentLocked(afterSeq, limit, level, source)
}

func (r *MemoryLogRing) recentLocked(afterSeq uint64, limit int, level, source string) []MemoryLogEntry {
	result := make([]MemoryLogEntry, 0, limit)
	start := (r.next - r.count + MemoryLogCapacity) % MemoryLogCapacity
	for i := 0; i < r.count; i++ {
		slot := &r.slots[(start+i)%MemoryLogCapacity]
		if slot.seq <= afterSeq || !logLevelMatches(slot.level, level) || source != "" && slot.source != source {
			continue
		}
		result = append(result, MemoryLogEntry{
			Seq:     slot.seq,
			Time:    slot.time,
			Level:   slot.level,
			Source:  slot.source,
			Message: string(slot.message[:slot.length]),
		})
		if len(result) > limit {
			result = result[len(result)-limit:]
		}
	}
	return result
}

// logLevelMatches treats the selected level as a minimum severity. An empty
// filter accepts every level, including trace.
func logLevelMatches(entryLevel, selectedLevel string) bool {
	if selectedLevel == "" {
		return true
	}
	ranks := map[string]int{"trace": 0, "debug": 1, "info": 2, "warn": 3, "error": 4}
	entryRank, entryOK := ranks[entryLevel]
	selectedRank, selectedOK := ranks[selectedLevel]
	return entryOK && selectedOK && entryRank >= selectedRank
}

func (r *MemoryLogRing) Stats() (count int, dropped uint64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.count, r.dropped
}

func (r *MemoryLogRing) Clear() {
	r.mu.Lock()
	r.next = 0
	r.count = 0
	r.dropped = 0
	r.mu.Unlock()
}

func (r *MemoryLogRing) Subscribe() (<-chan MemoryLogEntry, func()) {
	return r.SubscribeFiltered("", "")
}

// SubscribeFiltered subscribes to future entries matching the optional level
// and source filters. Filtering before the SSE layer prevents high-volume
// logs that are not displayed from reaching the browser.
func (r *MemoryLogRing) SubscribeFiltered(level, source string) (<-chan MemoryLogEntry, func()) {
	client, unsubscribe, _ := r.subscribeFilteredAfter(0, level, source)
	return client, unsubscribe
}

// SubscribeFilteredAfter atomically registers a subscriber and snapshots the
// matching entries after afterSeq. Registering before taking the snapshot
// prevents a pause/resume gap or duplicate entry between replay and live SSE.
func (r *MemoryLogRing) SubscribeFilteredAfter(afterSeq uint64, level, source string) (<-chan MemoryLogEntry, func(), []MemoryLogEntry) {
	return r.subscribeFilteredAfter(afterSeq, level, source)
}

func (r *MemoryLogRing) subscribeFilteredAfter(afterSeq uint64, level, source string) (<-chan MemoryLogEntry, func(), []MemoryLogEntry) {
	client := &memoryLogSubscriber{
		channel: make(chan MemoryLogEntry, 128),
		level:   strings.ToLower(strings.TrimSpace(level)),
		source:  strings.ToLower(strings.TrimSpace(source)),
	}
	r.mu.Lock()
	r.clients[client] = struct{}{}
	replay := r.recentLocked(afterSeq, MemoryLogCapacity, client.level, client.source)
	r.mu.Unlock()
	return client.channel, func() {
		r.mu.Lock()
		if _, ok := r.clients[client]; ok {
			delete(r.clients, client)
			close(client.channel)
		}
		r.mu.Unlock()
	}, replay
}

var globalMemoryLog = NewMemoryLogRing()

func GetMemoryLog() *MemoryLogRing { return globalMemoryLog }

// MemorySlogHandler copies all slog records to memory, independently of the
// file handler's configured level.
type MemorySlogHandler struct {
	hub    *MemoryLogRing
	attrs  []slog.Attr
	groups []string
}

func NewMemorySlogHandler(hub *MemoryLogRing) *MemorySlogHandler {
	return &MemorySlogHandler{hub: hub}
}

func (h *MemorySlogHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *MemorySlogHandler) Handle(_ context.Context, record slog.Record) error {
	parts := make([]string, 0, len(h.attrs)+record.NumAttrs())
	appendAttr := func(attr slog.Attr) bool {
		attr.Value = attr.Value.Resolve()
		if attr.Equal(slog.Attr{}) || attr.Key == "" {
			return true
		}
		key := attr.Key
		if len(h.groups) > 0 {
			key = strings.Join(append(append([]string{}, h.groups...), key), ".")
		}
		parts = append(parts, fmt.Sprintf("%s=%v", key, attr.Value.Any()))
		return true
	}
	for _, attr := range h.attrs {
		appendAttr(attr)
	}
	record.Attrs(appendAttr)
	message := record.Message
	if len(parts) > 0 {
		message += " " + strings.Join(parts, " ")
	}
	h.hub.Append(slogLevelName(record.Level), "panel", message)
	return nil
}

func (h *MemorySlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	copyAttrs := append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &MemorySlogHandler{hub: h.hub, attrs: copyAttrs, groups: append([]string{}, h.groups...)}
}

func (h *MemorySlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	groups := append(append([]string{}, h.groups...), name)
	return &MemorySlogHandler{hub: h.hub, attrs: append([]slog.Attr{}, h.attrs...), groups: groups}
}

func slogLevelName(level slog.Level) string {
	switch {
	case level <= slog.LevelDebug:
		return "debug"
	case level <= slog.LevelInfo:
		return "info"
	case level <= slog.LevelWarn:
		return "warn"
	default:
		return "error"
	}
}

type SingBoxMemoryWriter struct{ hub *MemoryLogRing }

func NewSingBoxMemoryWriter(hub *MemoryLogRing) *SingBoxMemoryWriter {
	return &SingBoxMemoryWriter{hub: hub}
}

func (w *SingBoxMemoryWriter) WriteMessage(level boxlog.Level, message string) {
	w.hub.Append(boxLevelName(level), "singbox", message)
}

func boxLevelName(level boxlog.Level) string {
	switch {
	case level <= boxlog.LevelError:
		return "error"
	case level <= boxlog.LevelWarn:
		return "warn"
	case level <= boxlog.LevelInfo:
		return "info"
	case level <= boxlog.LevelDebug:
		return "debug"
	default:
		return "trace"
	}
}
