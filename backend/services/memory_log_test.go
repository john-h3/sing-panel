package services

import (
	"strings"
	"testing"
	"time"

	boxlog "github.com/sagernet/sing-box/log"
)

func TestMemoryLogRingOverwritesOldestEntries(t *testing.T) {
	ring := NewMemoryLogRing()
	for i := 0; i < MemoryLogCapacity+2; i++ {
		ring.Append("info", "test", strings.Repeat("x", 1))
	}

	entries := ring.Recent(0, "", "")
	if len(entries) != MemoryLogCapacity {
		t.Fatalf("entry count = %d, want %d", len(entries), MemoryLogCapacity)
	}
	count, dropped := ring.Stats()
	if count != MemoryLogCapacity || dropped != 2 {
		t.Fatalf("stats = (%d, %d), want (%d, 2)", count, dropped, MemoryLogCapacity)
	}
	if entries[0].Seq != 3 || entries[len(entries)-1].Seq != MemoryLogCapacity+2 {
		t.Fatalf("sequence range = (%d, %d), want (3, %d)", entries[0].Seq, entries[len(entries)-1].Seq, MemoryLogCapacity+2)
	}
}

func TestBoxLevelName(t *testing.T) {
	tests := []struct {
		level boxlog.Level
		want  string
	}{
		{boxlog.LevelPanic, "error"},
		{boxlog.LevelFatal, "error"},
		{boxlog.LevelError, "error"},
		{boxlog.LevelWarn, "warn"},
		{boxlog.LevelInfo, "info"},
		{boxlog.LevelDebug, "debug"},
		{boxlog.LevelTrace, "trace"},
	}
	for _, test := range tests {
		if got := boxLevelName(test.level); got != test.want {
			t.Errorf("boxLevelName(%d) = %q, want %q", test.level, got, test.want)
		}
	}
}

func TestMemoryLogRingTruncatesAndFilters(t *testing.T) {
	ring := NewMemoryLogRing()
	longMessage := strings.Repeat("a", MemoryLogMaxBytes+100)
	ring.Append("trace", "singbox", "trace detail")
	ring.Append("debug", "panel", longMessage)
	ring.Append("error", "singbox", "failure")

	entries := ring.Recent(10, "error", "singbox")
	if len(entries) != 1 || entries[0].Message != "failure" {
		t.Fatalf("filtered entries = %+v, want one singbox error entry", entries)
	}

	entries = ring.Recent(10, "debug", "panel")
	if len(entries) != 1 || len(entries[0].Message) != MemoryLogMaxBytes {
		t.Fatalf("truncated message length = %d, want %d", len(entries[0].Message), MemoryLogMaxBytes)
	}

	for _, level := range []string{"debug", "info", "warn", "error"} {
		entries = ring.Recent(10, level, "singbox")
		if len(entries) != 1 || entries[0].Level != "error" {
			t.Fatalf("level %q filtered entries = %+v, want only error", level, entries)
		}
	}

	entries = ring.Recent(10, "", "singbox")
	if len(entries) != 2 || entries[0].Level != "trace" || entries[1].Level != "error" {
		t.Fatalf("unfiltered entries = %+v, want trace and error", entries)
	}
}

func TestMemoryLogRingSubscription(t *testing.T) {
	ring := NewMemoryLogRing()
	client, unsubscribe := ring.Subscribe()
	ring.Append("warn", "test", "hello")

	entry := <-client
	if entry.Message != "hello" || entry.Source != "test" {
		t.Fatalf("subscription entry = %+v", entry)
	}
	unsubscribe()
	select {
	case _, ok := <-client:
		if ok {
			t.Fatal("subscription channel should be closed")
		}
	default:
		t.Fatal("subscription channel was not closed")
	}
}

func TestMemoryLogRingSubscriptionFiltersTraceForSelectedLevels(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error"} {
		ring := NewMemoryLogRing()
		client, unsubscribe := ring.SubscribeFiltered(level, "singbox")
		ring.Append("trace", "singbox", "trace detail")
		select {
		case entry := <-client:
			t.Fatalf("level %q unexpectedly received trace entry = %+v", level, entry)
		default:
		}
		ring.Append("error", "singbox", "failure")
		select {
		case entry := <-client:
			if entry.Level != "error" {
				t.Fatalf("level %q subscription entry = %+v, want error entry", level, entry)
			}
		case <-time.After(time.Second):
			t.Fatalf("level %q did not receive error entry", level)
		}
		unsubscribe()
	}
}
