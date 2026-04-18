package main

import (
	"log"

	"docker-visual/internal/docker"
	"docker-visual/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	h := handlers.New(dockerClient)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.GET("/health", h.Health)

		api.GET("/containers", h.ListContainers)
		api.GET("/containers/:id", h.GetContainer)
		api.POST("/containers/:id/start", h.StartContainer)
		api.POST("/containers/:id/stop", h.StopContainer)
		api.DELETE("/containers/:id", h.RemoveContainer)

		api.GET("/networks", h.ListNetworks)
		api.GET("/networks/:id", h.GetNetwork)

		api.GET("/images", h.ListImages)

		api.GET("/volumes", h.ListVolumes)

		api.GET("/graph", h.GetGraphData)

		api.POST("/projects", h.CreateProject)
		api.GET("/projects", h.ListProjects)
		api.POST("/projects/:name/tunnel", h.CreateTunnel)
	}

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
