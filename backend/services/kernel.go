package services

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"sing_panel/models"
)

const (
	singBoxBinDir       = "./bin"
	singBoxBinName      = "sing-box"
	githubAPIURL        = "https://api.github.com/repos/SagerNet/sing-box/releases"
	cacheExpireDuration = 1 * time.Hour
)

type KernelService struct {
	mu             sync.RWMutex
	progress       models.DownloadProgress
	stopChan       chan struct{}
	versionsCache  []models.VersionInfo
	cacheMu        sync.RWMutex
	cacheUpdated   time.Time
	configService  *ConfigService
	db             *Database
	downloader     *Downloader
	apiClient      *http.Client
}

func NewKernelService(configService *ConfigService, db *Database) *KernelService {
	os.MkdirAll(singBoxBinDir, 0755)
	s := &KernelService{
		configService: configService,
		db:            db,
		downloader:    NewDownloader(),
		apiClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	// Try loading cache from database
	loaded, expired := s.loadCacheFromDB()
	if loaded {
		if expired {
			slog.Info("cache expired, refreshing in background")
			go s.refreshCache()
		} else {
			slog.Info("cache loaded from database")
		}
	} else {
		slog.Info("no cache found, fetching from GitHub")
		go s.refreshCache()
	}
	// Auto refresh every hour
	go s.startAutoRefresh()
	return s
}

func (s *KernelService) startAutoRefresh() {
	ticker := time.NewTicker(cacheExpireDuration)
	defer ticker.Stop()
	for range ticker.C {
		s.refreshCache()
	}
}

func (s *KernelService) refreshCache() {
	versions, err := s.fetchVersionsFromGitHub()
	if err != nil {
		slog.Error("cache refresh failed", "error", err)
		return
	}
	s.cacheMu.Lock()
	s.versionsCache = versions
	s.cacheUpdated = time.Now()
	s.cacheMu.Unlock()
	s.saveCacheToDB()
	slog.Info("cache refreshed", "time", s.cacheUpdated.Format("15:04:05"), "count", len(versions))
}

func (s *KernelService) loadCacheFromDB() (loaded bool, expired bool) {
	var cached struct {
		Updated  time.Time           `json:"updated"`
		Versions []models.VersionInfo `json:"versions"`
	}

	if err := s.db.Get("versions", "cache", &cached); err != nil {
		return false, false
	}

	expired = time.Since(cached.Updated) > cacheExpireDuration
	if expired {
		slog.Info("cache expired, using stale data", "age", time.Since(cached.Updated).Round(time.Second))
	}

	// Always load to memory, even if expired
	s.cacheMu.Lock()
	s.versionsCache = cached.Versions
	s.cacheUpdated = cached.Updated
	s.cacheMu.Unlock()

	return true, expired
}

func (s *KernelService) saveCacheToDB() {
	s.cacheMu.RLock()
	cached := struct {
		Updated  time.Time           `json:"updated"`
		Versions []models.VersionInfo `json:"versions"`
	}{
		Updated:  s.cacheUpdated,
		Versions: s.versionsCache,
	}
	s.cacheMu.RUnlock()

	if err := s.db.Put("versions", "cache", cached); err != nil {
		slog.Error("cache save failed", "error", err)
	}
}

// GetStatus returns the current kernel status
func (s *KernelService) GetStatus() models.KernelStatus {
	s.mu.RLock()
	progress := s.progress
	s.mu.RUnlock()

	var state models.KernelState
	if err := s.db.Get("state", "kernel", &state); err != nil {
		return models.KernelStatus{
			Path:      filepath.Join(singBoxBinDir, singBoxBinName),
			Active:    progress.Active,
			Progress:  progress.Progress,
			Status:    progress.Status,
			StatusMsg: progress.Error,
		}
	}

	return models.KernelStatus{
		Installed:    state.Installed,
		Version:      state.Version,
		Path:         state.Path,
		LastUpdated:  state.LastUpdated,
		DownloadType: state.DownloadType,
		Active:       progress.Active,
		Progress:     progress.Progress,
		Status:       progress.Status,
		StatusMsg:    progress.Error,
	}
}

// saveState saves kernel state to database
func (s *KernelService) saveState(state models.KernelState) error {
	return s.db.Put("state", "kernel", state)
}

// loadState loads kernel state from database
func (s *KernelService) loadState() (models.KernelState, error) {
	var state models.KernelState
	err := s.db.Get("state", "kernel", &state)
	return state, err
}

// savePID saves the process PID to state
func (s *KernelService) savePID(pid int) {
	s.db.UpdateKernelState(func(state *models.KernelState) {
		state.PID = pid
	})
}

// saveStartTime saves the process start time to state
func (s *KernelService) saveStartTime(startTime time.Time) {
	s.db.UpdateKernelState(func(state *models.KernelState) {
		state.StartTime = startTime
	})
}

// clearPID clears the PID from state
func (s *KernelService) clearPID() {
	s.db.UpdateKernelState(func(state *models.KernelState) {
		state.PID = 0
	})
}

// RefreshVersions manually refreshes the cache
func (s *KernelService) RefreshVersions() {
	s.refreshCache()
}

// GetCacheTime returns the last cache update time
func (s *KernelService) GetCacheTime() time.Time {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	return s.cacheUpdated
}

// GetVersions returns cached versions, fetching from GitHub if cache is empty
func (s *KernelService) GetVersions() ([]models.VersionInfo, error) {
	s.cacheMu.RLock()
	if len(s.versionsCache) > 0 {
		defer s.cacheMu.RUnlock()
		return s.versionsCache, nil
	}
	s.cacheMu.RUnlock()

	// Cache empty, fetch directly
	return s.fetchVersionsFromGitHub()
}

// fetchVersionsFromGitHub fetches available versions from GitHub
func (s *KernelService) fetchVersionsFromGitHub() ([]models.VersionInfo, error) {
	resp, err := s.apiClient.Get(githubAPIURL + "?per_page=20")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var releases []struct {
		TagName     string    `json:"tag_name"`
		Prerelease  bool      `json:"prerelease"`
		PublishedAt time.Time `json:"published_at"`
		Assets      []struct {
			Name          string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size          int64  `json:"size"`
			DownloadCount int    `json:"download_count"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode versions: %w", err)
	}

	var versions []models.VersionInfo
	for _, r := range releases {
		// Filter assets for current platform
		var assets []models.Asset
		for _, a := range r.Assets {
			if s.isRelevantAsset(a.Name) {
				assets = append(assets, models.Asset{
					Name:          a.Name,
					DownloadURL:   a.BrowserDownloadURL,
					Size:          a.Size,
					DownloadCount: a.DownloadCount,
				})
			}
		}

		if len(assets) > 0 {
			versions = append(versions, models.VersionInfo{
				Version:     strings.TrimPrefix(r.TagName, "v"),
				Tag:         r.TagName,
				Prerelease:  r.Prerelease,
				PublishedAt: r.PublishedAt,
				Assets:      assets,
			})
		}
	}

	return versions, nil
}

// isRelevantAsset checks if an asset is for the current platform
func (s *KernelService) isRelevantAsset(name string) bool {
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch to common names
	archMap := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
		"arm":   "armv7",
	}

	mappedArch, ok := archMap[arch]
	if !ok {
		return false
	}

	nameLower := strings.ToLower(name)
	return strings.Contains(nameLower, platform) &&
		strings.Contains(nameLower, mappedArch) &&
		strings.HasSuffix(nameLower, ".tar.gz")
}

// Download starts downloading a kernel version
func (s *KernelService) Download(req models.DownloadRequest) error {
	s.mu.Lock()
	if s.progress.Active {
		s.mu.Unlock()
		return fmt.Errorf("download already in progress")
	}
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	go s.doDownload(req)
	return nil
}

// StopDownload stops the current download
func (s *KernelService) StopDownload() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.progress.Active {
		return fmt.Errorf("no download in progress")
	}

	close(s.stopChan)
	return nil
}

// GetProgress returns the current download progress
func (s *KernelService) GetProgress() models.DownloadProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.progress
}

func (s *KernelService) doDownload(req models.DownloadRequest) {
	s.updateProgress(true, 0, "downloading", "", "")

	var downloadURL string
	var version string

	switch req.Type {
	case "latest":
		url, ver, err := s.getLatestVersion()
		if err != nil {
			s.updateProgress(false, 0, "failed", "", err.Error())
			return
		}
		downloadURL = url
		version = ver

	case "stable":
		url, ver, err := s.getStableVersion()
		if err != nil {
			s.updateProgress(false, 0, "failed", "", err.Error())
			return
		}
		downloadURL = url
		version = ver

	case "custom":
		if req.URL == "" {
			s.updateProgress(false, 0, "failed", "", "custom URL required")
			return
		}
		downloadURL = req.URL
		version = req.Version
		if version == "" {
			version = "custom"
		}

	default:
		s.updateProgress(false, 0, "failed", "", "invalid download type")
		return
	}

	s.updateProgress(true, 10, "downloading", version, "")

	// Download using the abstracted downloader
	tmpPath, err := s.downloader.DownloadToTemp(DownloadOptions{
		URL:          downloadURL,
		AccelerateURL: s.configService.Get().AccelerateDomain,
		StopChan:     s.stopChan,
		OnProgress: func(progress float64, speed int64) {
			// Map 0-100 to 10-60 for download phase
			s.updateProgress(true, 10+progress*0.5, "downloading", version, "")
		},
	})
	if err != nil {
		if err.Error() == "download cancelled" {
			s.updateProgress(false, 0, "failed", version, "download cancelled")
		} else {
			s.updateProgress(false, 0, "failed", version, err.Error())
		}
		return
	}
	defer os.Remove(tmpPath)

	s.updateProgress(true, 70, "downloading", version, "")

	// Extract tar.gz
	if err := s.extractArchive(tmpPath); err != nil {
		s.updateProgress(false, 0, "failed", version, err.Error())
		return
	}

	s.updateProgress(true, 90, "downloading", version, "")

	// Set executable permission
	binPath := filepath.Join(singBoxBinDir, singBoxBinName)
	if err := os.Chmod(binPath, 0755); err != nil {
		s.updateProgress(false, 0, "failed", version, err.Error())
		return
	}

	// Save state to database
	state := models.KernelState{
		Version:      version,
		Path:         binPath,
		Installed:    true,
		LastUpdated:  time.Now(),
		DownloadType: req.Type,
	}
	if err := s.saveState(state); err != nil {
		s.updateProgress(false, 0, "failed", version, err.Error())
		return
	}

	s.updateProgress(false, 100, "completed", version, "")
}

func (s *KernelService) updateProgress(active bool, progress float64, status, version, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.progress = models.DownloadProgress{
		Active:   active,
		Progress: progress,
		Status:   status,
		Version:  version,
		Error:    errMsg,
	}
}

func (s *KernelService) getLatestVersion() (string, string, error) {
	versions, err := s.GetVersions()
	if err != nil {
		return "", "", err
	}
	if len(versions) == 0 {
		return "", "", fmt.Errorf("no versions found")
	}

	v := versions[0]
	if len(v.Assets) == 0 {
		return "", "", fmt.Errorf("no assets found for version %s", v.Version)
	}

	return v.Assets[0].DownloadURL, v.Version, nil
}

func (s *KernelService) getStableVersion() (string, string, error) {
	versions, err := s.GetVersions()
	if err != nil {
		return "", "", err
	}

	for _, v := range versions {
		if !v.Prerelease && len(v.Assets) > 0 {
			return v.Assets[0].DownloadURL, v.Version, nil
		}
	}

	return "", "", fmt.Errorf("no stable version found")
}

func (s *KernelService) extractArchive(archivePath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Look for sing-box binary
		baseName := filepath.Base(header.Name)
		if baseName == singBoxBinName || baseName == singBoxBinName+".exe" {
			outPath := filepath.Join(singBoxBinDir, singBoxBinName)
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			return nil
		}
	}

	return fmt.Errorf("sing-box binary not found in archive")
}

// Remove removes the installed kernel
func (s *KernelService) Remove() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.progress.Active {
		return fmt.Errorf("cannot remove during download")
	}

	binPath := filepath.Join(singBoxBinDir, singBoxBinName)
	if err := os.Remove(binPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Clear state from database
	s.db.Delete("state", "kernel")
	return nil
}

// SwitchVersion switches to a specific version (re-download)
func (s *KernelService) SwitchVersion(version string) error {
	versions, err := s.GetVersions()
	if err != nil {
		return err
	}

	for _, v := range versions {
		if v.Version == version {
			if len(v.Assets) == 0 {
				return fmt.Errorf("no assets found for version %s", version)
			}
			return s.Download(models.DownloadRequest{
				Type:    "custom",
				Version: version,
				URL:     v.Assets[0].DownloadURL,
			})
		}
	}

	return fmt.Errorf("version %s not found", version)
}

// Verify checks if the installed binary is working
func (s *KernelService) Verify() (string, error) {
	binPath := filepath.Join(singBoxBinDir, singBoxBinName)
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return "", fmt.Errorf("binary not found")
	}

	cmd := exec.Command(binPath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run sing-box: %w", err)
	}

	return string(output), nil
}
