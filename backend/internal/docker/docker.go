package docker

import (
	"context"
	"fmt"

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

// NewService creates a new Docker service connected to the Docker daemon.
func NewService(cfg config.DockerConfig) (Service, error) {
	// TODO: Initialize actual Docker client using docker/docker SDK
	// client, err := client.NewClientWithOpts(client.FromEnv)
	return &stubService{socket: cfg.Local.Socket}, nil
}

type stubService struct {
	socket string
}

func (s *stubService) ListContainers(ctx context.Context, all bool) ([]models.Container, error) {
	// TODO: Implement using Docker SDK
	return []models.Container{}, nil
}

func (s *stubService) GetContainer(ctx context.Context, id string) (*models.ContainerDetail, error) {
	// TODO: Implement using Docker SDK
	return nil, fmt.Errorf("container %s not found", id)
}

func (s *stubService) RestartContainer(ctx context.Context, id string) error {
	// TODO: Implement using Docker SDK
	return fmt.Errorf("not implemented")
}

func (s *stubService) StopContainer(ctx context.Context, id string) error {
	// TODO: Implement using Docker SDK
	return fmt.Errorf("not implemented")
}

func (s *stubService) DeleteContainer(ctx context.Context, id string, force bool) error {
	// TODO: Implement using Docker SDK
	return fmt.Errorf("not implemented")
}
