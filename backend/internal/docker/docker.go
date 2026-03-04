package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

// Service defines the interface for Docker container operations.
type Service interface {
	ListContainers(ctx context.Context, all bool) ([]models.Container, error)
	GetContainer(ctx context.Context, id string) (*models.ContainerDetail, error)
	RestartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string) error
	DeleteContainer(ctx context.Context, id string, force bool) error
}

type dockerService struct {
	client *dockerclient.Client
}

// NewService creates a new Docker service connected to the Docker daemon.
func NewService(cfg config.DockerConfig) (Service, error) {
	opts := []dockerclient.Opt{
		dockerclient.WithAPIVersionNegotiation(),
	}
	if cfg.Local.Socket != "" {
		opts = append(opts, dockerclient.WithHost(cfg.Local.Socket))
	} else {
		opts = append(opts, dockerclient.FromEnv)
	}

	cli, err := dockerclient.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}

	return &dockerService{client: cli}, nil
}

func (s *dockerService) ListContainers(ctx context.Context, all bool) ([]models.Container, error) {
	containers, err := s.client.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	result := make([]models.Container, 0, len(containers))
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		ports := make([]models.ContainerPort, 0, len(c.Ports))
		for _, p := range c.Ports {
			ports = append(ports, models.ContainerPort{
				PrivatePort: int(p.PrivatePort),
				PublicPort:  int(p.PublicPort),
				Type:        p.Type,
			})
		}

		result = append(result, models.Container{
			ID:        c.ID[:12],
			Name:      name,
			Image:     c.Image,
			Status:    c.Status,
			State:     c.State,
			CreatedAt: time.Unix(c.Created, 0),
			Ports:     ports,
		})
	}
	return result, nil
}

func (s *dockerService) GetContainer(ctx context.Context, id string) (*models.ContainerDetail, error) {
	inspect, err := s.client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("inspecting container: %w", err)
	}

	name := strings.TrimPrefix(inspect.Name, "/")

	var ports []models.ContainerPort
	for port, bindings := range inspect.NetworkSettings.Ports {
		for _, b := range bindings {
			publicPort := 0
			fmt.Sscanf(b.HostPort, "%d", &publicPort)
			ports = append(ports, models.ContainerPort{
				PrivatePort: port.Int(),
				PublicPort:  publicPort,
				Type:        port.Proto(),
			})
		}
	}

	var env []string
	var cmd []string
	var workingDir string
	var exposedPorts []string
	if inspect.Config != nil {
		env = inspect.Config.Env
		cmd = inspect.Config.Cmd
		workingDir = inspect.Config.WorkingDir
		for p := range inspect.Config.ExposedPorts {
			exposedPorts = append(exposedPorts, string(p))
		}
	}

	mounts := make([]models.Mount, 0, len(inspect.Mounts))
	for _, m := range inspect.Mounts {
		mounts = append(mounts, models.Mount{
			Type:        string(m.Type),
			Source:      m.Source,
			Destination: m.Destination,
		})
	}

	var network models.NetworkInfo
	if inspect.NetworkSettings != nil {
		for _, n := range inspect.NetworkSettings.Networks {
			network = models.NetworkInfo{
				IPAddress:  n.IPAddress,
				Gateway:    n.Gateway,
				MACAddress: n.MacAddress,
			}
			break
		}
	}

	createdAt, _ := time.Parse(time.RFC3339Nano, inspect.Created)

	detail := &models.ContainerDetail{
		Container: models.Container{
			ID:        inspect.ID[:12],
			Name:      name,
			Image:     inspect.Config.Image,
			Status:    inspect.State.Status,
			State:     inspect.State.Status,
			CreatedAt: createdAt,
			Ports:     ports,
		},
		Config: models.ContainerConfig{
			Env:          env,
			Cmd:          cmd,
			WorkingDir:   workingDir,
			ExposedPorts: exposedPorts,
		},
		Mounts:  mounts,
		Network: network,
	}

	stats, err := s.getContainerStats(ctx, id)
	if err == nil {
		detail.Container.Stats = stats
	}

	return detail, nil
}

func (s *dockerService) getContainerStats(ctx context.Context, id string) (*models.ContainerStats, error) {
	resp, err := s.client.ContainerStats(ctx, id, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var statsJSON types.StatsJSON
	if err := json.Unmarshal(body, &statsJSON); err != nil {
		return nil, err
	}

	cpuDelta := float64(statsJSON.CPUStats.CPUUsage.TotalUsage - statsJSON.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(statsJSON.CPUStats.SystemUsage - statsJSON.PreCPUStats.SystemUsage)
	cpuPercent := 0.0
	if systemDelta > 0 && cpuDelta > 0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(statsJSON.CPUStats.OnlineCPUs) * 100.0
	}

	memUsage := int64(statsJSON.MemoryStats.Usage)
	memLimit := int64(statsJSON.MemoryStats.Limit)
	memPercent := 0.0
	if memLimit > 0 {
		memPercent = float64(memUsage) / float64(memLimit) * 100.0
	}

	var networkRx, networkTx int64
	for _, v := range statsJSON.Networks {
		networkRx += int64(v.RxBytes)
		networkTx += int64(v.TxBytes)
	}

	return &models.ContainerStats{
		CPUPercent:    cpuPercent,
		MemoryUsage:   memUsage,
		MemoryLimit:   memLimit,
		MemoryPercent: memPercent,
		NetworkRx:     networkRx,
		NetworkTx:     networkTx,
	}, nil
}

func (s *dockerService) RestartContainer(ctx context.Context, id string) error {
	timeout := 10
	return s.client.ContainerRestart(ctx, id, container.StopOptions{Timeout: &timeout})
}

func (s *dockerService) StopContainer(ctx context.Context, id string) error {
	timeout := 10
	return s.client.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout})
}

func (s *dockerService) DeleteContainer(ctx context.Context, id string, force bool) error {
	return s.client.ContainerRemove(ctx, id, container.RemoveOptions{Force: force})
}
