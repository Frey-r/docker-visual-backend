package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{cli: cli}, nil
}

func (c *Client) ListContainers(ctx context.Context) ([]types.Container, error) {
	return c.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
}

func (c *Client) GetContainer(ctx context.Context, id string) (types.ContainerJSON, error) {
	return c.cli.ContainerInspect(ctx, id)
}

func (c *Client) ListNetworks(ctx context.Context) ([]types.NetworkResource, error) {
	return c.cli.NetworkList(ctx, types.NetworkListOptions{})
}

func (c *Client) GetNetwork(ctx context.Context, id string) (types.NetworkResource, error) {
	return c.cli.NetworkInspect(ctx, id, types.NetworkInspectOptions{})
}

func (c *Client) ListImages(ctx context.Context) ([]types.ImageSummary, error) {
	return c.cli.ImageList(ctx, types.ImageListOptions{All: true})
}

func (c *Client) ListVolumes(ctx context.Context) ([]types.Volume, error) {
	return c.cli.VolumeList(ctx, types.VolumeListOptions{})
}

func (c *Client) StartContainer(ctx context.Context, id string) error {
	return c.cli.ContainerStart(ctx, id, types.ContainerStartOptions{})
}

func (c *Client) StopContainer(ctx context.Context, id string) error {
	return c.cli.ContainerStop(ctx, id, types.ContainerStopOptions{})
}

func (c *Client) RemoveContainer(ctx context.Context, id string, force bool) error {
	return c.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: force})
}

func (c *Client) GetStats(ctx context.Context, id string, stream bool) (<-chan types.StatsJSON, error) {
	return c.cli.ContainerStats(ctx, id, stream)
}

func (c *Client) Close() error {
	return c.cli.Close()
}
