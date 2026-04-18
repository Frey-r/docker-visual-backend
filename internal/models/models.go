package models



type Container struct {
	ID      string   `json:"id"`
	Names   []string `json:"names"`
	Image   string   `json:"image"`
	State   string   `json:"state"`
	Status  string   `json:"status"`
	Created int64    `json:"created"`
	Ports   []Port   `json:"ports"`
}

type Port struct {
	IP          string `json:"ip"`
	PrivatePort int    `json:"private_port"`
	PublicPort  int    `json:"public_port"`
	Type        string `json:"type"`
}

type Network struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Containers []NetworkEndpoint `json:"containers"`
}

type NetworkEndpoint struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	IPv4Address   string `json:"ipv4_address"`
	IPv6Address   string `json:"ipv6_address"`
	MacAddress    string `json:"mac_address"`
	ContainerID   string `json:"container_id"`
}

type Image struct {
	ID       string   `json:"id"`
	Tags     []string `json:"tags"`
	Size     int64    `json:"size"`
	Created  int64    `json:"created"`
	RepoTags []string `json:"repo_tags"`
}

type Volume struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	Labels     map[string]string `json:"labels"`
	CreatedAt  string            `json:"created_at"`
}

type Stats struct {
	CPU    CPUStats    `json:"cpu"`
	Memory MemoryStats `json:"memory"`
	IO     IOStats     `json:"io"`
	Net    NetStats    `json:"net"`
}

type CPUStats struct {
	UsagePercent float64 `json:"usage_percent"`
}

type MemoryStats struct {
	Usage    uint64 `json:"usage"`
	Limit    uint64 `json:"limit"`
	UsagePercent float64 `json:"usage_percent"`
}

type IOStats struct {
	Read  int64 `json:"read"`
	Write int64 `json:"write"`
}

type NetStats struct {
	RxBytes int64 `json:"rx_bytes"`
	TxBytes int64 `json:"tx_bytes"`
}

type GraphNode struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // container, network
	Label string `json:"label"`
	Data  any    `json:"data"`
}

type GraphLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"` // network_container, etc
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

type Project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	NetworkID  string `json:"network_id"`
	Containers int    `json:"containers"`
}

type TunnelRequest struct {
	Token string `json:"token"`
}
