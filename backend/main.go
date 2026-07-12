package main

import (
	"log/slog"
	"os"
	"sing_panel/handlers"
	"sing_panel/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const dbPath = "./data/sing-panel.db"

func main() {
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

	// Setup Gin router
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Serve static frontend files
	router.Static("/assets", "./frontend/dist/assets")
	router.StaticFile("/", "./frontend/dist/index.html")
	router.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	// API routes
	api := router.Group("/api")
	{
		kernelHandler := handlers.NewKernelHandler(kernelService)
		configHandler := handlers.NewConfigHandler(configService)
		singboxConfigHandler := handlers.NewSingBoxConfigHandler(singboxConfigService)
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
			kernel.DELETE("/", kernelHandler.Remove)
			kernel.POST("/switch", kernelHandler.SwitchVersion)
		}

		config := api.Group("/config")
		{
			config.GET("/", configHandler.Get)
			config.PUT("/", configHandler.Update)
		}

		singbox := api.Group("/singbox")
		{
			singbox.GET("/", singboxConfigHandler.GetConfig)
			singbox.GET("/export", singboxConfigHandler.ExportConfig)

			singbox.GET("/inbounds", singboxConfigHandler.GetInbounds)
			singbox.POST("/inbounds", singboxConfigHandler.AddInbound)
			singbox.PUT("/inbounds", singboxConfigHandler.UpdateInbound)
			singbox.DELETE("/inbounds/:id", singboxConfigHandler.DeleteInbound)

			singbox.GET("/outbounds", singboxConfigHandler.GetOutbounds)
			singbox.POST("/outbounds", singboxConfigHandler.AddOutbound)
			singbox.PUT("/outbounds", singboxConfigHandler.UpdateOutbound)
			singbox.DELETE("/outbounds/:id", singboxConfigHandler.DeleteOutbound)

			singbox.GET("/types/inbound", singboxConfigHandler.GetInboundTypes)
			singbox.GET("/types/outbound", singboxConfigHandler.GetOutboundTypes)
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
	}

	// Clash API Proxy - frontend calls this to access Clash API
	proxyHandler := handlers.NewProxyHandler()
	router.Any("/clash_api/*path", proxyHandler.GenericProxy)

	slog.Info("server starting", "port", 8080)
	if err := router.Run(":8080"); err != nil {
		slog.Error("server failed", "error", err)
	}
}
