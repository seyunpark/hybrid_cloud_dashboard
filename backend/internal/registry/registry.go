package registry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	dockerclient "github.com/docker/docker/client"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
)

// Service defines the interface for container registry operations.
type Service interface {
	PushImage(ctx context.Context, sourceImage, targetImage string) error
	TagImage(ctx context.Context, source, target string) error
}

type registryService struct {
	url      string
	username string
	password string
	client   *dockerclient.Client
}

// NewService creates a new registry service with the given configuration.
func NewService(cfg config.RegistryConfig) (Service, error) {
	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating docker client for registry: %w", err)
	}

	return &registryService{
		url:      cfg.Default.URL,
		username: cfg.Default.Username,
		password: cfg.Default.Password,
		client:   cli,
	}, nil
}

func (s *registryService) TagImage(ctx context.Context, source, target string) error {
	return s.client.ImageTag(ctx, source, target)
}

func (s *registryService) PushImage(ctx context.Context, sourceImage, targetImage string) error {
	if err := s.TagImage(ctx, sourceImage, targetImage); err != nil {
		return fmt.Errorf("tagging image: %w", err)
	}

	authConfig := registry.AuthConfig{
		Username:      s.username,
		Password:      s.password,
		ServerAddress: s.url,
	}
	encodedAuth, err := encodeAuthConfig(authConfig)
	if err != nil {
		return fmt.Errorf("encoding auth: %w", err)
	}

	pushResp, err := s.client.ImagePush(ctx, targetImage, types.ImagePushOptions{
		RegistryAuth: encodedAuth,
	})
	if err != nil {
		return fmt.Errorf("pushing image: %w", err)
	}
	defer pushResp.Close()

	// Read push output to completion
	decoder := json.NewDecoder(pushResp)
	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("reading push output: %w", err)
		}
		if errMsg, ok := msg["error"]; ok {
			return fmt.Errorf("push error: %v", errMsg)
		}
		if status, ok := msg["status"]; ok {
			slog.Debug("push progress", "status", status)
		}
	}

	return nil
}

func encodeAuthConfig(authConfig registry.AuthConfig) (string, error) {
	encoded, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(encoded), nil
}
