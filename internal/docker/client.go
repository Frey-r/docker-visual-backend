package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"docker-visual/internal/models"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// DockerClient defines the interface for Docker operations.
// This enables mocking in tests.
type DockerClient interface {
	ListContainers(ctx context.Context) ([]types.Container, error)
	GetContainer(ctx context.Context, id string) (types.ContainerJSON, error)
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string) error
	RemoveContainer(ctx context.Context, id string, force bool) error
	ListNetworks(ctx context.Context) ([]network.Inspect, error)
	GetNetwork(ctx context.Context, id string) (network.Inspect, error)
	ListImages(ctx context.Context) ([]image.Summary, error)
	ListVolumes(ctx context.Context) ([]*volume.Volume, error)
	CreateProjectNetwork(ctx context.Context, name string) (string, error)
	RunCloudflaredContainer(ctx context.Context, projectName, networkID, token string) error
	BuildImage(ctx context.Context, buildContextPath, imageName string) error
	CreateAndStartContainer(ctx context.Context, imageName, networkID, projectName string) error
	CreateContainerFromImage(ctx context.Context, req models.CreateContainerRequest) (string, string, error)
	PullImage(ctx context.Context, imageName string) error
	Ping(ctx context.Context) error
	Close() error
}

// Client wraps the Docker SDK client and implements DockerClient.
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client from environment configuration.
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{cli: cli}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	return err
}

func (c *Client) ListContainers(ctx context.Context) ([]types.Container, error) {
	return c.cli.ContainerList(ctx, container.ListOptions{All: true})
}

func (c *Client) GetContainer(ctx context.Context, id string) (types.ContainerJSON, error) {
	return c.cli.ContainerInspect(ctx, id)
}

func (c *Client) ListNetworks(ctx context.Context) ([]network.Inspect, error) {
	return c.cli.NetworkList(ctx, network.ListOptions{})
}

func (c *Client) GetNetwork(ctx context.Context, id string) (network.Inspect, error) {
	return c.cli.NetworkInspect(ctx, id, network.InspectOptions{})
}

func (c *Client) ListImages(ctx context.Context) ([]image.Summary, error) {
	return c.cli.ImageList(ctx, image.ListOptions{All: true})
}

func (c *Client) ListVolumes(ctx context.Context) ([]*volume.Volume, error) {
	res, err := c.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}
	return res.Volumes, nil
}

func (c *Client) StartContainer(ctx context.Context, id string) error {
	return c.cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (c *Client) StopContainer(ctx context.Context, id string) error {
	return c.cli.ContainerStop(ctx, id, container.StopOptions{})
}

func (c *Client) RemoveContainer(ctx context.Context, id string, force bool) error {
	return c.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: force})
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) CreateProjectNetwork(ctx context.Context, name string) (string, error) {
	resp, err := c.cli.NetworkCreate(ctx, name, network.CreateOptions{
		Driver: "bridge",
		Labels: map[string]string{
			"docker-dashboard.project": "true",
			"docker-dashboard.name":    name,
		},
	})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) RunCloudflaredContainer(ctx context.Context, projectName string, networkID string, token string) error {
	// 1. Pull the image and wait for the download to complete.
	reader, err := c.cli.ImagePull(ctx, "cloudflare/cloudflared:latest", image.PullOptions{})
	if err != nil {
		return err
	}
	// Drain the reader fully to ensure the image is downloaded before proceeding.
	if _, err := io.Copy(io.Discard, reader); err != nil {
		reader.Close()
		return err
	}
	reader.Close()

	// 2. Create the container
	resp, err := c.cli.ContainerCreate(ctx, &container.Config{
		Image: "cloudflare/cloudflared:latest",
		Cmd:   []string{"tunnel", "--no-autoupdate", "run", "--token", token},
		Labels: map[string]string{
			"docker-dashboard.project": projectName,
			"docker-dashboard.service": "cloudflared",
		},
	}, &container.HostConfig{
		NetworkMode: container.NetworkMode(networkID),
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}, nil, nil, "cloudflared-"+projectName)
	if err != nil {
		return err
	}

	// 3. Start the container
	return c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
}

func (c *Client) BuildImage(ctx context.Context, buildContextPath string, imageName string) error {
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", imageName, buildContextPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) CreateAndStartContainer(ctx context.Context, imageName string, networkID string, projectName string) error {
	resp, err := c.cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Labels: map[string]string{
			"docker-dashboard.project": projectName,
		},
	}, &container.HostConfig{
		NetworkMode:     container.NetworkMode(networkID),
		PublishAllPorts: true,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}, nil, nil, projectName+"-app")
	if err != nil {
		return err
	}

	return c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
}

func (c *Client) PullImage(ctx context.Context, imageName string) error {
	reader, err := c.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(io.Discard, reader)
	return err
}

func (c *Client) CreateContainerFromImage(ctx context.Context, req models.CreateContainerRequest) (string, string, error) {
	// Build environment variables slice
	var envVars []string
	for key, value := range req.Env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	// Build port bindings
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for _, p := range req.Ports {
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		containerPort := nat.Port(fmt.Sprintf("%d/%s", p.ContainerPort, proto))
		exposedPorts[containerPort] = struct{}{}
		portBindings[containerPort] = []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", p.HostPort)},
		}
	}

	// Build volume mounts
	var mounts []mount.Mount
	for _, v := range req.Volumes {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   v.HostPath,
			Target:   v.ContainerPath,
			ReadOnly: v.ReadOnly,
		})
	}

	// Determine restart policy
	restartPolicy := container.RestartPolicy{Name: container.RestartPolicyDisabled}
	switch req.RestartPolicy {
	case "always":
		restartPolicy = container.RestartPolicy{Name: container.RestartPolicyAlways}
	case "unless-stopped":
		restartPolicy = container.RestartPolicy{Name: container.RestartPolicyUnlessStopped}
	case "on-failure":
		restartPolicy = container.RestartPolicy{Name: container.RestartPolicyOnFailure}
	}

	// Container config
	config := &container.Config{
		Image:        req.Image,
		Env:          envVars,
		ExposedPorts: exposedPorts,
		Labels: map[string]string{
			"docker-dashboard.managed": "true",
		},
	}

	// Host config
	hostConfig := &container.HostConfig{
		PortBindings:  portBindings,
		Mounts:        mounts,
		RestartPolicy: restartPolicy,
	}

	// Network config
	if req.NetworkID != "" {
		hostConfig.NetworkMode = container.NetworkMode(req.NetworkID)
	}

	// Create container
	resp, err := c.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, req.Name)
	if err != nil {
		return "", "", err
	}

	// Start container
	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return resp.ID, req.Name, err
	}

	return resp.ID, req.Name, nil
}
