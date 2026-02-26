package registry

import (
	"context"
	"fmt"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
)

// Service defines the interface for container registry operations.
type Service interface {
	PushImage(ctx context.Context, sourceImage, targetImage string) error
	TagImage(ctx context.Context, source, target string) error
}

// NewService creates a new registry service with the given configuration.
func NewService(cfg config.RegistryConfig) (Service, error) {
	// TODO: Initialize registry client based on registry type (Docker Hub, ECR, ACR, GCR)
	return &stubService{
		url: cfg.Default.URL,
	}, nil
}

type stubService struct {
	url string
}

func (s *stubService) PushImage(ctx context.Context, sourceImage, targetImage string) error {
	// TODO: Implement image push to container registry
	// 1. Tag image with registry URL
	// 2. Push using Docker SDK
	return fmt.Errorf("not implemented")
}

func (s *stubService) TagImage(ctx context.Context, source, target string) error {
	// TODO: Implement image tagging
	return fmt.Errorf("not implemented")
}
