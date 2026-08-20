package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultLogPath       = "/var/log/sing-panel.log"
	defaultLogMaxBytes   = 100 * 1024 * 1024
	defaultLogMaxBackups = 7
	defaultLogMaxAge     = 14 * 24 * time.Hour
)

// RollingFileWriter writes to a file and rotates it before it exceeds maxBytes.
// Rotated files are gzip-compressed and retained according to maxBackups/maxAge.
type RollingFileWriter struct {
	mu          sync.Mutex
	path        string
	maxBytes    int64
	maxBackups  int
	maxAge      time.Duration
	file        *os.File
	currentSize int64
	compressWG  sync.WaitGroup
}

func NewRollingFileWriter(path string, maxBytes int64, maxBackups int, maxAge time.Duration) (*RollingFileWriter, error) {
	if path == "" {
		return nil, fmt.Errorf("log path is empty")
	}
	if maxBytes <= 0 {
		return nil, fmt.Errorf("log max size must be positive")
	}
	if maxBackups < 0 {
		return nil, fmt.Errorf("log max backups cannot be negative")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("stat log file: %w", err)
	}

	w := &RollingFileWriter{
		path:        path,
		maxBytes:    maxBytes,
		maxBackups:  maxBackups,
		maxAge:      maxAge,
		file:        file,
		currentSize: stat.Size(),
	}
	if w.currentSize >= w.maxBytes {
		if err := w.rotateLocked(); err != nil {
			if w.file != nil {
				_ = w.file.Close()
			}
			return nil, fmt.Errorf("rotate existing log file: %w", err)
		}
	}
	w.cleanupLocked()
	return w, nil
}

func (w *RollingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(p) == 0 {
		return 0, nil
	}
	if w.currentSize > 0 && w.currentSize+int64(len(p)) > w.maxBytes {
		if err := w.rotateLocked(); err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(p)
	w.currentSize += int64(n)
	return n, err
}

func (w *RollingFileWriter) Close() error {
	w.mu.Lock()
	if w.file == nil {
		w.mu.Unlock()
		w.compressWG.Wait()
		return nil
	}
	err := w.file.Close()
	w.file = nil
	w.mu.Unlock()
	// Compression is intentionally asynchronous during rotation, but must be
	// allowed to finish before the process exits so archives are not left as
	// partial .gz.tmp files.
	w.compressWG.Wait()
	return err
}

func (w *RollingFileWriter) rotateLocked() error {
	if w.file == nil {
		return fmt.Errorf("log file is closed")
	}
	if err := w.file.Close(); err != nil {
		return err
	}
	w.file = nil
	reopen := func() error {
		file, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		stat, err := file.Stat()
		if err != nil {
			_ = file.Close()
			return err
		}
		w.file = file
		w.currentSize = stat.Size()
		return nil
	}

	archiveBase := fmt.Sprintf("%s-%s", w.path, time.Now().UTC().Format("20060102-150405.000000000"))
	archivePath := archiveBase
	for i := 1; ; i++ {
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			break
		}
		archivePath = fmt.Sprintf("%s-%d", archiveBase, i)
	}
	if err := os.Rename(w.path, archivePath); err != nil {
		// Keep the writer usable if rotation cannot rename the active file.
		// This can happen with unusual filesystem permissions or mounts.
		if reopenErr := reopen(); reopenErr != nil {
			return fmt.Errorf("rename log file: %w (reopen: %v)", err, reopenErr)
		}
		return err
	}

	// Compress asynchronously. Renaming the active file and reopening it must
	// stay fast even when the previous log is hundreds of megabytes large.
	// A failed compression leaves the uncompressed archive in place, so a
	// rotation failure never destroys the previous log contents.
	compressedPath := archivePath + ".gz"
	if err := reopen(); err != nil {
		return err
	}
	w.compressWG.Add(1)
	go func() {
		defer w.compressWG.Done()
		if err := compressFile(archivePath, compressedPath); err == nil {
			_ = os.Remove(archivePath)
		}
		w.mu.Lock()
		w.cleanupLocked()
		w.mu.Unlock()
	}()
	w.cleanupLocked()
	return nil
}

func compressFile(sourcePath, targetPath string) error {
	tempPath := targetPath + ".tmp"
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	zipWriter := gzip.NewWriter(target)
	if _, err = io.Copy(zipWriter, source); err != nil {
		_ = zipWriter.Close()
		_ = target.Close()
		return err
	}
	if err = zipWriter.Close(); err != nil {
		_ = target.Close()
		return err
	}
	if err = target.Close(); err != nil {
		return err
	}
	if err = os.Rename(tempPath, targetPath); err != nil {
		return err
	}
	removeTemp = false
	return nil
}

func (w *RollingFileWriter) cleanupLocked() {
	pattern := w.path + "-*.gz"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	sort.Strings(matches)
	cutoff := time.Now().Add(-w.maxAge)
	for i, path := range matches {
		remove := w.maxBackups >= 0 && len(matches)-i > w.maxBackups
		if !remove && w.maxAge > 0 {
			if stat, statErr := os.Stat(path); statErr == nil && stat.ModTime().Before(cutoff) {
				remove = true
			}
		}
		if remove {
			_ = os.Remove(path)
		}
	}
}

// setupProcessLogging routes the process-level stdout and stderr through the
// application-owned rolling file. sing-box, Gin, slog, and other libraries
// that use the standard streams consequently share the same rotation policy.
func setupProcessLogging(path string, maxBytes int64) (func() error, error) {
	rolling, err := NewRollingFileWriter(path, maxBytes, defaultLogMaxBackups, defaultLogMaxAge)
	if err != nil {
		return nil, err
	}

	reader, writer, err := os.Pipe()
	if err != nil {
		_ = rolling.Close()
		return nil, fmt.Errorf("create log pipe: %w", err)
	}

	// Keep a single pipe for both streams so output ordering is preserved as
	// much as possible. Redirect the actual file descriptors as well as the Go
	// variables: OpenRC/systemd may have already bound fd 1/2 to a file before
	// starting this process.
	redirect, err := redirectStandardStreams(writer)
	if err != nil {
		_ = writer.Close()
		_ = reader.Close()
		_ = rolling.Close()
		return nil, fmt.Errorf("redirect standard streams: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		_, copyErr := io.Copy(rolling, bufio.NewReader(reader))
		_ = reader.Close()
		done <- copyErr
	}()

	var closeOnce sync.Once
	var closeErr error
	cleanup := func() error {
		closeOnce.Do(func() {
			// Close both duplicated pipe write descriptors. This causes io.Copy to
			// drain and then receive EOF.
			closeErr = redirect.closeWriters()
			copyErr := <-done
			if closeErr == nil {
				closeErr = copyErr
			}
			if err := rolling.Close(); closeErr == nil {
				closeErr = err
			}
			if err := redirect.restore(); closeErr == nil {
				closeErr = err
			}
		})
		return closeErr
	}
	return cleanup, nil
}

func logPathFromEnv(defaultPath string) string {
	if path := strings.TrimSpace(os.Getenv("SING_PANEL_LOG_FILE")); path != "" {
		return path
	}
	return defaultPath
}

func logMaxBytesFromEnv(defaultBytes int64) (int64, error) {
	value := strings.TrimSpace(os.Getenv("SING_PANEL_LOG_MAX_MB"))
	if value == "" {
		return defaultBytes, nil
	}
	megabytes, err := strconv.ParseInt(value, 10, 64)
	if err != nil || megabytes <= 0 {
		return 0, fmt.Errorf("SING_PANEL_LOG_MAX_MB must be a positive integer")
	}
	return megabytes * 1024 * 1024, nil
}
