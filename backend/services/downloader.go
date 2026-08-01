package services

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// DownloadProgressFunc is called during download to report progress
// progress: 0-100, speed: bytes per second
type DownloadProgressFunc func(progress float64, speed int64)

// DownloadOptions configures the download behavior
type DownloadOptions struct {
	URL              string
	DestPath         string
	AccelerateURL    string   // If set, use this as prefix instead of direct URL
	AccelerateDomains []string // Domains that should be accelerated
	OnProgress       DownloadProgressFunc
	StopChan         <-chan struct{}
}

// Downloader handles file downloads with accelerate domain support
type Downloader struct {
	httpClient *http.Client
}

func NewDownloader() *Downloader {
	return &Downloader{
		httpClient: &http.Client{
			Timeout: 30 * time.Minute,
		},
	}
}

// GetAccelerateURL applies the accelerate domain to a GitHub URL
// If accelerateDomain is empty, returns the original URL
// Otherwise returns: accelerateDomain + "/" + originalURL
func GetAccelerateURL(rawURL, accelerateDomain string) string {
	if accelerateDomain == "" {
		return rawURL
	}
	domain := accelerateDomain
	if len(domain) > 0 && domain[len(domain)-1] == '/' {
		domain = domain[:len(domain)-1]
	}
	return domain + "/" + rawURL
}

// ShouldAccelerate checks if the URL's domain matches any of the configured accelerate domains.
// If accelerateDomains is empty, all URLs are eligible for acceleration (legacy behavior).
func ShouldAccelerate(rawURL string, accelerateDomains []string) bool {
	if len(accelerateDomains) == 0 {
		return true // no filter configured, accelerate all
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	for _, pattern := range accelerateDomains {
		p := strings.ToLower(strings.TrimSpace(pattern))
		if p == "" {
			continue
		}
		// exact match or suffix match (e.g. "github.com" matches "raw.githubusercontent.com")
		if host == p || strings.HasSuffix(host, "."+p) {
			return true
		}
	}
	return false
}

// Download downloads a file from the given URL to destPath
// It applies the accelerate domain if configured and URL matches the filter
func (d *Downloader) Download(opts DownloadOptions) error {
	// Determine final URL
	finalURL := opts.URL
	if opts.AccelerateURL != "" && ShouldAccelerate(opts.URL, opts.AccelerateDomains) {
		finalURL = GetAccelerateURL(opts.URL, opts.AccelerateURL)
		slog.Info("using accelerate domain", "original", opts.URL, "accelerated", finalURL)
	} else {
		slog.Info("direct download", "url", opts.URL)
	}

	// Create temp file first, rename on success
	tmpFile, err := os.CreateTemp("", "download-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Cleanup on failure
	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	// Start download
	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpFile.Close()
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength
	var downloaded int64
	startTime := time.Now()

	buf := make([]byte, 32*1024)
	for {
		// Check stop signal
		if opts.StopChan != nil {
			select {
			case <-opts.StopChan:
				tmpFile.Close()
				return fmt.Errorf("download cancelled")
			default:
			}
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
				tmpFile.Close()
				return fmt.Errorf("write: %w", writeErr)
			}
			downloaded += int64(n)

			// Report progress
			if opts.OnProgress != nil && totalSize > 0 {
				elapsed := time.Since(startTime).Seconds()
				var speed int64
				if elapsed > 0 {
					speed = int64(float64(downloaded) / elapsed)
				}
				progress := float64(downloaded) / float64(totalSize) * 100
				opts.OnProgress(progress, speed)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			tmpFile.Close()
			return fmt.Errorf("read: %w", err)
		}
	}

	tmpFile.Close()

	// Rename temp file to destination
	if err := os.Rename(tmpPath, opts.DestPath); err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	success = true
	return nil
}

// DownloadToTemp downloads to a temp file and returns the path
// Caller is responsible for cleaning up the file
func (d *Downloader) DownloadToTemp(opts DownloadOptions) (string, error) {
	tmpFile, err := os.CreateTemp("", "download-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	opts.DestPath = tmpPath

	if err := d.Download(opts); err != nil {
		os.Remove(tmpPath)
		return "", err
	}

	return tmpPath, nil
}
