package handlers

import (
	"docker-visual/internal/docker"
	"docker-visual/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	docker *docker.Client
}

func New(d *docker.Client) *Handler {
	return &Handler{docker: d}
}

func (h *Handler) ListContainers(c *gin.Context) {
	containers, err := h.docker.ListContainers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]models.Container, len(containers))
	for i, c := range containers {
		ports := make([]models.Port, len(c.Ports))
		for j, p := range c.Ports {
			ports[j] = models.Port{
				IP:          p.IP,
				PrivatePort: int(p.PrivatePort),
				PublicPort:  int(p.PublicPort),
				Type:        p.Type,
			}
		}
		names := make([]string, len(c.Names))
		for j, n := range c.Names {
			names[j] = n
		}
		result[i] = models.Container{
			ID:      c.ID,
			Names:   names,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: c.Created,
			Ports:   ports,
		}
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetContainer(c *gin.Context) {
	id := c.Param("id")
	info, err := h.docker.GetContainer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (h *Handler) StartContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.docker.StartContainer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Container started"})
}

func (h *Handler) StopContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.docker.StopContainer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Container stopped"})
}

func (h *Handler) RemoveContainer(c *gin.Context) {
	id := c.Param("id")
	force := c.Query("force") == "true"
	if err := h.docker.RemoveContainer(c.Request.Context(), id, force); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Container removed"})
}

func (h *Handler) ListNetworks(c *gin.Context) {
	networks, err := h.docker.ListNetworks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]models.Network, len(networks))
	for i, n := range networks {
		containers := make([]models.NetworkEndpoint, len(n.Containers))
		for j, cn := range n.Containers {
			containers[j] = models.NetworkEndpoint{
				ID:          cn.Name,
				Name:        cn.Name,
				IPv4Address: cn.IPv4Address,
				IPv6Address: cn.IPv6Address,
				MacAddress:  cn.MacAddress,
				ContainerID: cn.ContainerID,
			}
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

func (h *Handler) GetNetwork(c *gin.Context) {
	id := c.Param("id")
	network, err := h.docker.GetNetwork(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, network)
}

func (h *Handler) ListImages(c *gin.Context) {
	images, err := h.docker.ListImages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]models.Image, len(images))
	for i, img := range images {
		result[i] = models.Image{
			ID:      img.ID,
			Size:    img.Size,
			Created: img.Created,
			RepoTags: img.RepoTags,
		}
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListVolumes(c *gin.Context) {
	volumes, err := h.docker.ListVolumes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func (h *Handler) GetGraphData(c *gin.Context) {
	containers, err := h.docker.ListContainers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	networks, err := h.docker.ListNetworks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
