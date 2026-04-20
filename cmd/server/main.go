package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"docker-visual/internal/config"
	"docker-visual/internal/docker"
	"docker-visual/internal/handlers"
	"docker-visual/internal/jobs"
	"docker-visual/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup structured logging
	logLevel := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	// Connect to Docker
	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	// Verify Docker connectivity
	if err := dockerClient.Ping(context.Background()); err != nil {
		log.Fatalf("Docker engine is not reachable: %v", err)
	}
	slog.Info("connected to Docker engine")

	// Create dependencies
	tracker := jobs.NewTracker()
	h := handlers.New(dockerClient, cfg, tracker)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Routes
	api := r.Group("/api")
	{
		// Public
		api.GET("/health", h.Health)

		// Protected (requires API key when configured)
		protected := api.Group("")
		protected.Use(middleware.APIKeyAuth(cfg.APIKey))
		{
			protected.GET("/containers", h.ListContainers)
			protected.GET("/containers/:id", h.GetContainer)
			protected.POST("/containers/:id/start", h.StartContainer)
			protected.POST("/containers/:id/stop", h.StopContainer)
			protected.DELETE("/containers/:id", h.RemoveContainer)

			protected.GET("/networks", h.ListNetworks)
			protected.GET("/networks/:id", h.GetNetwork)

			protected.GET("/images", h.ListImages)

			protected.GET("/volumes", h.ListVolumes)

			protected.GET("/graph", h.GetGraphData)

			protected.POST("/projects", h.CreateProject)
			protected.GET("/projects", h.ListProjects)
			protected.POST("/projects/:name/tunnel", h.CreateTunnel)

			protected.GET("/deploy/status/:name", h.GetDeployStatus)
			protected.GET("/deploy/jobs", h.ListDeployJobs)
		}
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	slog.Info("server stopped")
}
