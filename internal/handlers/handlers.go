package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/exec"

	"docker-visual/internal/config"
	"docker-visual/internal/docker"
	"docker-visual/internal/jobs"
	"docker-visual/internal/models"
	"docker-visual/internal/validate"

	"github.com/gin-gonic/gin"
)

// Handler holds dependencies for all HTTP handlers.
type Handler struct {
	docker  docker.DockerClient
	cfg     *config.Config
	jobs    *jobs.Tracker
	logger  *slog.Logger
}

// New creates a new Handler with all required dependencies.
func New(d docker.DockerClient, cfg *config.Config, tracker *jobs.Tracker) *Handler {
	return &Handler{
		docker: d,
		cfg:    cfg,
		jobs:   tracker,
		logger: slog.Default(),
	}
}

// Health checks the server and Docker engine connectivity.
func (h *Handler) Health(c *gin.Context) {
	if err := h.docker.Ping(c.Request.Context()); err != nil {
		h.logger.Error("docker engine unreachable", "error", err)
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error: "docker engine unreachable",
			Code:  "DOCKER_UNAVAILABLE",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ListContainers returns all containers on the Docker host.
func (h *Handler) ListContainers(c *gin.Context) {
	containers, err := h.docker.ListContainers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}

	result := make([]models.Container, len(containers))
	for i, ct := range containers {
		ports := make([]models.Port, len(ct.Ports))
		for j, p := range ct.Ports {
			ports[j] = models.Port{
				IP:          p.IP,
				PrivatePort: int(p.PrivatePort),
				PublicPort:  int(p.PublicPort),
				Type:        p.Type,
			}
		}
		result[i] = models.Container{
			ID:      ct.ID,
			Names:   ct.Names,
			Image:   ct.Image,
			State:   ct.State,
			Status:  ct.Status,
			Created: ct.Created,
			Ports:   ports,
		}
	}
	c.JSON(http.StatusOK, result)
}

// GetContainer returns detailed info for a single container.
func (h *Handler) GetContainer(c *gin.Context) {
	id := c.Param("id")
	if err := validate.ContainerID(id); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_INPUT"})
		return
	}

	info, err := h.docker.GetContainer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: err.Error(), Code: "NOT_FOUND"})
		return
	}
	c.JSON(http.StatusOK, info)
}

// StartContainer starts a stopped container.
func (h *Handler) StartContainer(c *gin.Context) {
	id := c.Param("id")
	if err := validate.ContainerID(id); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_INPUT"})
		return
	}

	if err := h.docker.StartContainer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}
	h.logger.Info("container started", "id", id)
	c.JSON(http.StatusOK, gin.H{"message": "Container started"})
}

// StopContainer stops a running container.
func (h *Handler) StopContainer(c *gin.Context) {
	id := c.Param("id")
	if err := validate.ContainerID(id); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_INPUT"})
		return
	}

	if err := h.docker.StopContainer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}
	h.logger.Info("container stopped", "id", id)
	c.JSON(http.StatusOK, gin.H{"message": "Container stopped"})
}

// RemoveContainer deletes a container. Use ?force=true to force removal.
func (h *Handler) RemoveContainer(c *gin.Context) {
	id := c.Param("id")
	if err := validate.ContainerID(id); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_INPUT"})
		return
	}

	force := c.Query("force") == "true"
	if err := h.docker.RemoveContainer(c.Request.Context(), id, force); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}
	h.logger.Info("container removed", "id", id, "force", force)
	c.JSON(http.StatusOK, gin.H{"message": "Container removed"})
}

// ListNetworks returns all Docker networks.
func (h *Handler) ListNetworks(c *gin.Context) {
	networks, err := h.docker.ListNetworks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}

	result := make([]models.Network, len(networks))
	for i, n := range networks {
		containers := make([]models.NetworkEndpoint, 0, len(n.Containers))
		for containerID, cn := range n.Containers {
			containers = append(containers, models.NetworkEndpoint{
				ID:          cn.Name,
				Name:        cn.Name,
				IPv4Address: cn.IPv4Address,
				IPv6Address: cn.IPv6Address,
				MacAddress:  cn.MacAddress,
				ContainerID: containerID,
			})
		}
		result[i] = models.Network{
			ID:         n.ID,
			Name:       n.Name,
			Driver:     n.Driver,
			Scope:      n.Scope,
			Containers: containers,
		}
	}
	c.JSON(http.StatusOK, result)
}

// GetNetwork returns detailed info for a single network.
func (h *Handler) GetNetwork(c *gin.Context) {
	id := c.Param("id")
	net, err := h.docker.GetNetwork(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: err.Error(), Code: "NOT_FOUND"})
		return
	}
	c.JSON(http.StatusOK, net)
}

// ListImages returns all Docker images.
func (h *Handler) ListImages(c *gin.Context) {
	images, err := h.docker.ListImages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}

	result := make([]models.Image, len(images))
	for i, img := range images {
		result[i] = models.Image{
			ID:       img.ID,
			Size:     img.Size,
			Created:  img.Created,
			RepoTags: img.RepoTags,
		}
	}
	c.JSON(http.StatusOK, result)
}

// ListVolumes returns all Docker volumes.
func (h *Handler) ListVolumes(c *gin.Context) {
	volumes, err := h.docker.ListVolumes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}

	result := make([]models.Volume, len(volumes))
	for i, v := range volumes {
		result[i] = models.Volume{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			Labels:     v.Labels,
			CreatedAt:  v.CreatedAt,
		}
	}
	c.JSON(http.StatusOK, result)
}

// GetGraphData returns the infrastructure topology as nodes and links.
func (h *Handler) GetGraphData(c *gin.Context) {
	containers, err := h.docker.ListContainers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}

	networks, err := h.docker.ListNetworks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}

	nodes := []models.GraphNode{}
	links := []models.GraphLink{}

	// Add network nodes
	for _, n := range networks {
		nodes = append(nodes, models.GraphNode{
			ID:    n.ID,
			Type:  "network",
			Label: n.Name,
			Data:  n,
		})
	}

	// Add container nodes and links
	for _, ct := range containers {
		containerName := ct.Names[0]
		if len(containerName) > 0 && containerName[0] == '/' {
			containerName = containerName[1:]
		}

		nodes = append(nodes, models.GraphNode{
			ID:    ct.ID,
			Type:  "container",
			Label: containerName,
			Data:  ct,
		})

		// Link container to its networks
		for _, net := range ct.NetworkSettings.Networks {
			links = append(links, models.GraphLink{
				Source: ct.ID,
				Target: net.NetworkID,
				Type:   "network_container",
			})
		}
	}

	c.JSON(http.StatusOK, models.GraphData{Nodes: nodes, Links: links})
}

// CreateProject creates a new project: network + optional git clone/build/deploy.
func (h *Handler) CreateProject(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		GitURL string `json:"gitUrl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_INPUT"})
		return
	}

	// Validate project name
	if err := validate.ProjectName(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_PROJECT_NAME"})
		return
	}

	// Validate git URL
	if err := validate.GitURL(req.GitURL); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_GIT_URL"})
		return
	}

	networkID, err := h.docker.CreateProjectNetwork(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}

	if req.GitURL != "" {
		// Validate workspace path
		workspace, err := validate.WorkspacePath(h.cfg.WorkspacePath, req.Name)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_PATH"})
			return
		}

		// Register the job for tracking
		h.jobs.Create(req.Name, req.GitURL, networkID)

		go h.runDeploy(req.Name, req.GitURL, networkID, workspace)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project created", "network_id": networkID, "name": req.Name})
}

// runDeploy performs the git clone → docker build → container start pipeline
// in the background, updating job status at each step.
func (h *Handler) runDeploy(projectName, gitURL, networkID, workspace string) {
	ctx := context.Background()

	// Clone
	h.jobs.UpdateStatus(projectName, jobs.StatusCloning)
	h.logger.Info("cloning repository", "project", projectName, "url", gitURL)
	os.MkdirAll(workspace, 0755)

	if err := cloneRepo(ctx, gitURL, workspace); err != nil {
		h.logger.Error("clone failed", "project", projectName, "error", err)
		h.jobs.SetError(projectName, err)
		return
	}

	// Build
	h.jobs.UpdateStatus(projectName, jobs.StatusBuilding)
	imageName := "project-" + projectName
	h.logger.Info("building image", "project", projectName, "image", imageName)
	if err := h.docker.BuildImage(ctx, workspace, imageName); err != nil {
		h.logger.Error("build failed", "project", projectName, "error", err)
		h.jobs.SetError(projectName, err)
		return
	}

	// Start
	h.jobs.UpdateStatus(projectName, jobs.StatusStarting)
	h.logger.Info("starting container", "project", projectName)
	if err := h.docker.CreateAndStartContainer(ctx, imageName, networkID, projectName); err != nil {
		h.logger.Error("start failed", "project", projectName, "error", err)
		h.jobs.SetError(projectName, err)
		return
	}

	h.jobs.UpdateStatus(projectName, jobs.StatusDone)
	h.logger.Info("deploy completed", "project", projectName)
}

// cloneRepo runs git clone in a subprocess.
func cloneRepo(ctx context.Context, gitURL, workspace string) error {
	cmd := execCommand(ctx, "git", "clone", gitURL, workspace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// execCommand wraps exec.CommandContext for testability.
var execCommand = defaultExecCommand

func defaultExecCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// ListProjects returns projects identified by Docker network labels.
func (h *Handler) ListProjects(c *gin.Context) {
	networks, err := h.docker.ListNetworks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}

	projects := []models.Project{}
	for _, n := range networks {
		if n.Labels["docker-dashboard.project"] == "true" {
			projects = append(projects, models.Project{
				ID:         n.ID,
				Name:       n.Labels["docker-dashboard.name"],
				NetworkID:  n.ID,
				Containers: len(n.Containers),
			})
		}
	}
	c.JSON(http.StatusOK, projects)
}

// GetDeployStatus returns the current status of a deploy job.
func (h *Handler) GetDeployStatus(c *gin.Context) {
	name := c.Param("name")
	job := h.jobs.Get(name)
	if job == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "no deploy job found", Code: "NOT_FOUND"})
		return
	}
	c.JSON(http.StatusOK, job)
}

// ListDeployJobs returns all deploy jobs.
func (h *Handler) ListDeployJobs(c *gin.Context) {
	c.JSON(http.StatusOK, h.jobs.List())
}

// CreateTunnel creates a Cloudflare tunnel for a project.
func (h *Handler) CreateTunnel(c *gin.Context) {
	projectName := c.Param("name")

	if err := validate.ProjectName(projectName); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_INPUT"})
		return
	}

	// Find project network
	networks, err := h.docker.ListNetworks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}

	var networkID string
	for _, n := range networks {
		if n.Labels["docker-dashboard.project"] == "true" && n.Labels["docker-dashboard.name"] == projectName {
			networkID = n.ID
			break
		}
	}

	if networkID == "" {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "project network not found", Code: "NOT_FOUND"})
		return
	}

	var req models.TunnelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "INVALID_INPUT"})
		return
	}

	if err := h.docker.RunCloudflaredContainer(c.Request.Context(), projectName, networkID, req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error(), Code: "DOCKER_ERROR"})
		return
	}

	h.logger.Info("cloudflare tunnel started", "project", projectName)
	c.JSON(http.StatusOK, gin.H{"message": "Cloudflared tunnel started"})
}
