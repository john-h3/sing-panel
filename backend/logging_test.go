package main

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRollingFileWriterRotatesAndCompresses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sing-panel.log")
	writer, err := NewRollingFileWriter(path, 10, 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("1234567890")); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("abc")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != "abc" {
		t.Fatalf("current log = %q, want %q", current, "abc")
	}

	archives, err := filepath.Glob(path + "-*.gz")
	if err != nil || len(archives) != 1 {
		t.Fatalf("compressed archives = %v, want one", archives)
	}
	archive, err := os.Open(archives[0])
	if err != nil {
		t.Fatal(err)
	}
	decompressor, err := gzip.NewReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	content, err := io.ReadAll(decompressor)
	if err != nil {
		t.Fatal(err)
	}
	_ = decompressor.Close()
	_ = archive.Close()
	if string(content) != "1234567890" {
		t.Fatalf("archive = %q, want %q", content, "1234567890")
	}
}

func TestLogMaxBytesFromEnv(t *testing.T) {
	t.Setenv("SING_PANEL_LOG_MAX_MB", "2")
	maxBytes, err := logMaxBytesFromEnv(100)
	if err != nil {
		t.Fatal(err)
	}
	if maxBytes != 2*1024*1024 {
		t.Fatalf("max bytes = %d, want %d", maxBytes, 2*1024*1024)
	}

	t.Setenv("SING_PANEL_LOG_MAX_MB", "invalid")
	if _, err := logMaxBytesFromEnv(100); err == nil {
		t.Fatal("expected invalid environment value to fail")
	}
}

func TestRollingFileWriterAgeCleanup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sing-panel.log")
	writer, err := NewRollingFileWriter(path, 1, 10, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("a")); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("b")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	archives, err := filepath.Glob(path + "-*.gz")
	if err != nil || len(archives) != 1 {
		t.Fatalf("compressed archives = %v, want one", archives)
	}
}
