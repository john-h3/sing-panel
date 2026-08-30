package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sing-panel/handlers"
	"sing-panel/services"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

//go:embed frontend/dist/*
var frontendDist embed.FS

const defaultDataDir = "./data"

const usage = `Sing Box Panel - sing-box 管理面板

Usage:
  sing-panel <command> [options]

Commands:
  run         启动服务
  install     安装为系统服务 (支持 systemd/openrc)
  uninstall   卸载系统服务

Options:
  --listen <addr>      监听地址 (default :8080)
  --data-dir <dir>     数据目录（配置文件所在目录，default ./data）
  -h, --help           显示帮助信息

Examples:
  sing-panel run                       # 默认监听 :8080
  sing-panel run --listen :3000        # 监听 :3000
  sing-panel run --data-dir /etc/sing-panel   # 指定数据目录
  sing-panel install                   # 安装为系统服务
  sing-panel install --listen :3000    # 安装并指定端口
  sing-panel uninstall                 # 卸载系统服务
`

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help" {
		fmt.Print(usage)
		return
	}

	switch os.Args[1] {
	case "run":
		// handled below
	case "install":
		installCmd := flag.NewFlagSet("install", flag.ExitOnError)
		listen := installCmd.String("listen", ":8080", "监听地址")
		dataDir := installCmd.String("data-dir", defaultDataDir, "数据目录")
		installCmd.Parse(os.Args[2:])
		cmdInstall(*listen, *dataDir)
		return
	case "uninstall":
		cmdUninstall()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		fmt.Print(usage)
		os.Exit(1)
	}

	// Parse run subcommand flags
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	listen := runCmd.String("listen", ":8080", "监听地址")
	dataDir := runCmd.String("data-dir", defaultDataDir, "数据目录")
	runCmd.Parse(os.Args[2:])

	logPath := logPathFromEnv(defaultLogPath)
	logMaxBytes, err := logMaxBytesFromEnv(defaultLogMaxBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "日志配置错误: %v\n", err)
		return
	}
	cleanupLogging, err := setupProcessLogging(logPath, logMaxBytes)
	if err != nil {
		// Local development commonly runs without permission to write /var/log.
		// Keep the service usable by falling back to its data directory; deployed
		// root/OpenRC installations continue to use /var/log by default.
		fallbackPath := filepath.Join(*dataDir, "sing-panel.log")
		cleanupLogging, err = setupProcessLogging(fallbackPath, logMaxBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "初始化应用日志失败: %v\n", err)
			return
		}
		logPath = fallbackPath
	}
	defer func() {
		if err := cleanupLogging(); err != nil {
			fmt.Fprintf(os.Stderr, "关闭应用日志失败: %v\n", err)
		}
	}()

	// Setup structured logging
	fileHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: &services.PanelLogLevel,
	})
	memoryHandler := services.NewMemorySlogHandler(services.GetMemoryLog())
	logger := slog.New(&fanoutHandler{handlers: []slog.Handler{fileHandler, memoryHandler}})
	slog.SetDefault(logger)
	slog.Info("application logging initialized", "path", logPath, "max_bytes", logMaxBytes)

	// Initialize database
	dbPath := filepath.Join(*dataDir, "sing-panel.db")
	db, err := services.NewDatabase(dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err, "path", dbPath)
		return
	}
	defer db.Close()
	slog.Info("database opened", "path", dbPath)

	// Initialize services
	configService := services.NewConfigService(db)
	kernelService := services.NewKernelService(db)
	singboxConfigService := services.NewSingBoxConfigService(db, configService)
	statsService := services.NewStatsService(db)
	processService := services.NewProcessService(db, singboxConfigService, kernelService, statsService, *dataDir)

	// Mark kernel as installed (embedded mode)
	kernelService.SetInstalled(true)

	// Start tree cache refresh loop
	singboxConfigService.StartTreeRefreshLoop()

	// Setup Gin router
	// Gin initializes these package-level writers before main runs, so update
	// them explicitly after the process output has been redirected to the
	// application-owned rolling log pipe.
	gin.DefaultWriter = &memoryLogWriter{writer: os.Stdout, level: slog.LevelInfo, source: "gin"}
	gin.DefaultErrorWriter = &memoryLogWriter{writer: os.Stderr, level: slog.LevelError, source: "gin"}
	router := gin.Default()
	router.RedirectTrailingSlash = false

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Sync-Token"},
		AllowCredentials: true,
	}))

	// Enable gzip compression for responses. Static assets under /assets are
	// served from pre-compressed .br/.gz build artifacts instead (see
	// precompressedFileServer below), so they are excluded here to avoid
	// compressing already-compressed content.
	router.Use(gzip.Gzip(gzip.DefaultCompression,
		gzip.WithExcludedPathsRegexs([]string{`^/assets/`})))

	// Serve embedded frontend files
	distFS, err := fs.Sub(frontendDist, "frontend/dist")
	if err != nil {
		slog.Error("failed to create sub filesystem", "error", err)
		return
	}

	// Serve assets (preferring pre-compressed .br/.gz build artifacts)
	assetServer := precompressedFileServer(http.FS(distFS))
	router.GET("/assets/*filepath", func(c *gin.Context) {
		assetServer.ServeHTTP(c.Writer, c.Request)
	})

	// SPA fallback - serve index.html for all non-API routes
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Try to serve the file directly from embedded FS
		if f, err := distFS.Open(strings.TrimPrefix(path, "/")); err == nil {
			f.Close()
			assetServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		// Fallback to index.html for SPA routing
		if !strings.HasPrefix(path, "/api") {
			f, err := distFS.Open("index.html")
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			defer f.Close()
			c.Data(http.StatusOK, "text/html", nil)
			io.Copy(c.Writer, f)
		}
	})

	// API routes
	api := router.Group("/api")
	{
		kernelHandler := handlers.NewKernelHandler(kernelService)
		configHandler := handlers.NewConfigHandler(configService)
		singboxConfigHandler := handlers.NewSingBoxConfigHandler(singboxConfigService, configService)
		processHandler := handlers.NewProcessHandler(processService)
		// Health is exposed at the panel root so load balancers and external
		// probes can check it without an API prefix.
		router.GET("/health", processHandler.Health)
		statsHandler := handlers.NewStatsHandler(statsService)

		// Multi-instance management
		instanceService := services.NewInstanceService(db, configService, processService, kernelService)
		instanceHandler := handlers.NewInstanceHandler(instanceService)
		panelHandler := handlers.NewPanelHandler(instanceService)

		systemHandler := handlers.NewSystemHandler()

		kernel := api.Group("/kernel")
		{
			kernel.GET("/status", kernelHandler.GetStatus)
			kernel.GET("/system", kernelHandler.GetSystemInfo)
			kernel.GET("/monitor", kernelHandler.GetMonitor)
		}

		config := api.Group("/config")
		{
			config.GET("", configHandler.Get)
			config.PUT("", configHandler.Update)
		}

		singbox := api.Group("/singbox")
		{
			singbox.GET("", singboxConfigHandler.GetConfig)
			singbox.GET("/export", singboxConfigHandler.ExportConfig)

			singbox.GET("/inbounds", singboxConfigHandler.GetInbounds)
			singbox.POST("/inbounds", singboxConfigHandler.AddInbound)
			singbox.PUT("/inbounds", singboxConfigHandler.UpdateInbound)
			singbox.DELETE("/inbounds/:id", singboxConfigHandler.DeleteInbound)

			singbox.GET("/outbounds", singboxConfigHandler.GetOutbounds)
			singbox.POST("/outbounds", singboxConfigHandler.AddOutbound)
			singbox.PUT("/outbounds", singboxConfigHandler.UpdateOutbound)
			singbox.DELETE("/outbounds/:id", singboxConfigHandler.DeleteOutbound)

			singbox.GET("/rulesets", singboxConfigHandler.GetRulesets)
			singbox.POST("/rulesets", singboxConfigHandler.AddRuleset)
			singbox.POST("/rulesets/batch", singboxConfigHandler.AddRulesets)
			singbox.PUT("/rulesets", singboxConfigHandler.UpdateRuleset)
			singbox.DELETE("/rulesets/:id", singboxConfigHandler.DeleteRuleset)
			singbox.POST("/rulesets/delete", singboxConfigHandler.DeleteRulesets)

			singbox.GET("/route-rules", singboxConfigHandler.GetRouteRules)
			singbox.POST("/route-rules", singboxConfigHandler.AddRouteRule)
			singbox.PUT("/route-rules", singboxConfigHandler.UpdateRouteRule)
			singbox.DELETE("/route-rules/:id", singboxConfigHandler.DeleteRouteRule)
			singbox.POST("/route-rules/batch-update", singboxConfigHandler.BatchUpdateRouteRules)
			singbox.POST("/route-rules/batch-delete", singboxConfigHandler.BatchDeleteRouteRules)
			singbox.POST("/route-rules/reorder", singboxConfigHandler.ReorderRouteRules)

			singbox.GET("/route-config", singboxConfigHandler.GetRouteConfig)
			singbox.PUT("/route-config", singboxConfigHandler.UpdateRouteConfig)

			singbox.GET("/dns", singboxConfigHandler.GetDNS)
			singbox.PUT("/dns", singboxConfigHandler.UpdateDNS)

			singbox.GET("/services", singboxConfigHandler.GetServices)
			singbox.POST("/services", singboxConfigHandler.AddService)
			singbox.PUT("/services", singboxConfigHandler.UpdateService)
			singbox.DELETE("/services/:id", singboxConfigHandler.DeleteService)

			singbox.GET("/http-clients", singboxConfigHandler.GetHTTPClients)
			singbox.POST("/http-clients", singboxConfigHandler.AddHTTPClient)
			singbox.PUT("/http-clients", singboxConfigHandler.UpdateHTTPClient)
			singbox.DELETE("/http-clients/:id", singboxConfigHandler.DeleteHTTPClient)

			singbox.GET("/experimental", singboxConfigHandler.GetExperimental)
			singbox.PUT("/experimental", singboxConfigHandler.UpdateExperimental)

			singbox.GET("/types/inbound", singboxConfigHandler.GetInboundTypes)
			singbox.GET("/types/outbound", singboxConfigHandler.GetOutboundTypes)
			singbox.GET("/network-interfaces", singboxConfigHandler.GetNetworkInterfaces)
			singbox.GET("/geo-tree", singboxConfigHandler.FetchGeoTree)
			singbox.GET("/common-ruleset-tree", singboxConfigHandler.FetchCommonRulesetTree)
			singbox.POST("/import", singboxConfigHandler.ImportLink)
		}

		process := api.Group("/process")
		{
			process.GET("/status", processHandler.GetStatus)
			process.GET("/config", processHandler.GetRuntimeConfig)
			process.POST("/start", processHandler.Start)
			process.POST("/stop", processHandler.Stop)
			process.POST("/restart", processHandler.Restart)
			process.POST("/reset-dashboard", processHandler.ResetDashboard)
		}

		stats := api.Group("/stats")
		{
			stats.GET("/service", statsHandler.GetServiceInfo)
		}

		system := api.Group("/system")
		{
			system.GET("/init", systemHandler.GetInitSystem)
			system.POST("/restart-service", systemHandler.RestartService)
			system.POST("/reboot-machine", systemHandler.RebootMachine)
		}

		instances := api.Group("/instances")
		{
			instances.GET("", instanceHandler.List)
			instances.POST("", instanceHandler.Create)
			instances.PUT("/:id", instanceHandler.Update)
			instances.DELETE("/:id", instanceHandler.Delete)
			instances.GET("/status", instanceHandler.CheckAll)
			instances.GET("/:id/status", instanceHandler.GetStatus)
			instances.POST("/:id/sync", instanceHandler.Sync)
			instances.POST("/sync-all", instanceHandler.SyncAll)
			instances.GET("/local-info", instanceHandler.LocalInfo)
			instances.PUT("/sync-token", instanceHandler.SetSyncToken)
			instances.GET("/:id/diff", instanceHandler.GetConfigDiff)
		}

		dbHandler := handlers.NewDatabaseHandler(db, configService)
		dbGroup := api.Group("/db")
		{
			dbGroup.GET("/buckets", dbHandler.ListBuckets)
			dbGroup.GET("/keys", dbHandler.ListKeys)
			dbGroup.GET("/value", dbHandler.GetValue)
			dbGroup.PUT("/value", dbHandler.PutValue)
			dbGroup.DELETE("/value", dbHandler.DeleteKey)
			dbGroup.DELETE("/bucket", dbHandler.DeleteBucket)
		}

		// Endpoints used for cross-panel management. When a sync token is
		// configured on this panel they additionally require the
		// X-Sync-Token header (checked per request, so changes take effect
		// without restarting).
		syncGuard := syncTokenGuard(instanceService)
		api.GET("/panel/info", syncGuard, panelHandler.Info)
		api.GET("/db/export", syncGuard, dbHandler.Export)
		api.POST("/db/import", syncGuard, dbHandler.Import)

		logHandler := handlers.NewLogHandler(services.GetMemoryLog())
		api.GET("/logs", logHandler.List)
		api.GET("/logs/stream", logHandler.Stream)
		api.DELETE("/logs", logHandler.Clear)
	}

	if configService.Get().AutoStartKernel {
		go func() {
			if err := processService.Start(); err != nil {
				slog.Error("failed to auto-start embedded sing-box", "error", err)
				return
			}
			slog.Info("embedded sing-box auto-started")
		}()
	}

	slog.Info("server starting", "listen", *listen)
	fmt.Printf("Sing Box Panel 已启动: http://%s\n", *listen)

	// Start server in a goroutine
	srv := &http.Server{Addr: *listen, Handler: router}
	// Long-lived SSE log streams never end on their own and would make
	// srv.Shutdown always run into its deadline. Close them up front.
	srv.RegisterOnShutdown(func() {
		services.GetMemoryLog().CloseAllSubscribers()
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server...")

	// Clean up iptables before exit
	processService.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}
	slog.Info("server exited")
}

// syncTokenGuard optionally requires the X-Sync-Token header to match the
// configured sync token for cross-panel sync endpoints. It reads the current
// token on every request, so enabling or disabling token protection takes
// effect immediately.
func syncTokenGuard(instanceService *services.InstanceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := instanceService.SyncToken()
		if token == "" {
			c.Next()
			return
		}
		if c.GetHeader("X-Sync-Token") != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid sync token"})
			return
		}
		c.Next()
	}
}

// precompressedFileServer serves static files, preferring the pre-compressed
// .br or .gz variant produced at build time when the client advertises support.
// Falls back to the plain file for clients without compression support.
//
// Vite content-hashes all asset filenames, so assets never change in place and
// can be cached immutably by browsers.
func precompressedFileServer(fs http.FileSystem) http.Handler {
	fileServer := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptEncoding := r.Header.Get("Accept-Encoding")
		ext, encoding := "", ""
		switch {
		case strings.Contains(acceptEncoding, "br"):
			ext, encoding = ".br", "br"
		case strings.Contains(acceptEncoding, "gzip"):
			ext, encoding = ".gz", "gzip"
		}

		if ext != "" {
			name := strings.TrimPrefix(r.URL.Path, "/")
			if f, err := fs.Open(name + ext); err == nil {
				if stat, err := f.Stat(); err == nil && !stat.IsDir() {
					defer f.Close()
					w.Header().Set("Content-Encoding", encoding)
					w.Header().Add("Vary", "Accept-Encoding")
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
					http.ServeContent(w, r, name, stat.ModTime(), f)
					return
				}
				f.Close()
			}
		}

		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		fileServer.ServeHTTP(w, r)
	})
}
