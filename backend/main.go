package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"sing_panel/handlers"
	"sing_panel/services"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed frontend/dist/*
var frontendDist embed.FS

const dbPath = "./data/sing-panel.db"

const usage = `Sing Box Panel - sing-box 管理面板

Usage:
  sing-panel <command> [options]

Commands:
  run         启动服务
  install     安装为系统服务 (支持 systemd/openrc)
  uninstall   卸载系统服务

Options:
  --listen <addr>   监听地址 (default :8080)
  -h, --help        显示帮助信息

Examples:
  sing-panel run                       # 默认监听 :8080
  sing-panel run --listen :3000        # 监听 :3000
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
		installCmd.Parse(os.Args[2:])
		cmdInstall(*listen)
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
	runCmd.Parse(os.Args[2:])

	// Setup structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Initialize database
	db, err := services.NewDatabase(dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		return
	}
	defer db.Close()

	// Initialize services
	configService := services.NewConfigService(db)
	kernelService := services.NewKernelService(configService, db)
	singboxConfigService := services.NewSingBoxConfigService(db)
	statsService := services.NewStatsService(db)
	processService := services.NewProcessService(db, singboxConfigService, kernelService, statsService)

	// Start tree cache refresh loop
	singboxConfigService.StartTreeRefreshLoop()

	// Setup Gin router
	router := gin.Default()
	router.RedirectTrailingSlash = false

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Serve embedded frontend files
	distFS, err := fs.Sub(frontendDist, "frontend/dist")
	if err != nil {
		slog.Error("failed to create sub filesystem", "error", err)
		return
	}

	// Serve assets
	fileServer := http.FileServer(http.FS(distFS))
	router.GET("/assets/*filepath", func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// SPA fallback - serve index.html for all non-API routes
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Try to serve the file directly from embedded FS
		if f, err := distFS.Open(strings.TrimPrefix(path, "/")); err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
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
		statsHandler := handlers.NewStatsHandler(statsService)

		kernel := api.Group("/kernel")
		{
			kernel.GET("/status", kernelHandler.GetStatus)
			kernel.GET("/system", kernelHandler.GetSystemInfo)
			kernel.GET("/versions", kernelHandler.GetVersions)
			kernel.POST("/versions/refresh", kernelHandler.RefreshVersions)
			kernel.POST("/download", kernelHandler.Download)
			kernel.POST("/stop", kernelHandler.StopDownload)
			kernel.DELETE("", kernelHandler.Remove)
			kernel.POST("/switch", kernelHandler.SwitchVersion)
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
			process.POST("/start", processHandler.Start)
			process.POST("/stop", processHandler.Stop)
			process.POST("/restart", processHandler.Restart)
		}

		stats := api.Group("/stats")
		{
			stats.GET("/service", statsHandler.GetServiceInfo)
		}

		dbHandler := handlers.NewDatabaseHandler(db)
		dbGroup := api.Group("/db")
		{
			dbGroup.GET("/buckets", dbHandler.ListBuckets)
			dbGroup.GET("/keys", dbHandler.ListKeys)
			dbGroup.GET("/value", dbHandler.GetValue)
			dbGroup.PUT("/value", dbHandler.PutValue)
			dbGroup.DELETE("/value", dbHandler.DeleteKey)
		}
	}

	slog.Info("server starting", "listen", *listen)
	fmt.Printf("Sing Box Panel 已启动: http://%s\n", *listen)
	if err := router.Run(*listen); err != nil {
		slog.Error("server failed", "error", err)
	}
}
