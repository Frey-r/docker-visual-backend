package docker

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
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
