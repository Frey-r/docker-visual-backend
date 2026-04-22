package models

// Container is the API response model for a Docker container.
type Container struct {
	ID      string   `json:"id"`
	Names   []string `json:"names"`
	Image   string   `json:"image"`
	State   string   `json:"state"`
	Status  string   `json:"status"`
	Created int64    `json:"created"`
	Ports   []Port   `json:"ports"`
}

// Port represents a container port mapping.
type Port struct {
	IP          string `json:"ip"`
	PrivatePort int    `json:"private_port"`
	PublicPort  int    `json:"public_port"`
	Type        string `json:"type"`
}

// Network is the API response model for a Docker network.
type Network struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Containers []NetworkEndpoint `json:"containers"`
}

// NetworkEndpoint represents a container attached to a network.
type NetworkEndpoint struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IPv4Address string `json:"ipv4_address"`
	IPv6Address string `json:"ipv6_address"`
	MacAddress  string `json:"mac_address"`
	ContainerID string `json:"container_id"`
}

// Image is the API response model for a Docker image.
type Image struct {
	ID       string   `json:"id"`
	Size     int64    `json:"size"`
	Created  int64    `json:"created"`
	RepoTags []string `json:"repo_tags"`
}

// Volume is the API response model for a Docker volume.
type Volume struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	Labels     map[string]string `json:"labels"`
	CreatedAt  string            `json:"created_at"`
}

// GraphNode represents a node in the infrastructure topology graph.
type GraphNode struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // "container" or "network"
	Label string `json:"label"`
	Data  any    `json:"data"`
}

// GraphLink represents an edge in the infrastructure topology graph.
type GraphLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"` // "network_container", etc.
}

// GraphData is the full topology graph response.
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

// Project represents a managed project (identified by Docker network labels).
type Project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	NetworkID  string `json:"network_id"`
	Containers int    `json:"containers"`
}

// TunnelRequest is the request body for creating a Cloudflare tunnel.
type TunnelRequest struct {
	Token string `json:"token" binding:"required"`
}

// ErrorResponse is the standard API error envelope.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// PortMapping represents a port mapping for container creation.
type PortMapping struct {
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"` // tcp or udp
}

// VolumeMapping represents a volume mount for container creation.
type VolumeMapping struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	ReadOnly      bool   `json:"read_only"`
}

// CreateContainerRequest is the request body for creating a container from an image.
type CreateContainerRequest struct {
	Image         string            `json:"image" binding:"required"`
	Name          string            `json:"name"`
	Env           map[string]string `json:"env"`
	Ports         []PortMapping     `json:"ports"`
	Volumes       []VolumeMapping   `json:"volumes"`
	NetworkID     string            `json:"network_id"`
	RestartPolicy string            `json:"restart_policy"` // no, always, unless-stopped, on-failure
}

// CreateContainerResponse is returned after container creation.
type CreateContainerResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}
