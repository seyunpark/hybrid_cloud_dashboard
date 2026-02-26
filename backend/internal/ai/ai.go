package ai

import (
	"context"
	"fmt"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

// ContainerInfo holds Docker container information used for AI manifest generation.
type ContainerInfo struct {
	Name        string
	Image       string
	ImageTag    string
	EnvVars     map[string]string
	Ports       []int
	Volumes     []string
	Command     []string
	WorkingDir  string
	CPUUsage    string
	MemoryUsage string
	NetworkMode string
}

// Service defines the interface for AI-based manifest generation.
type Service interface {
	GenerateManifest(ctx context.Context, info ContainerInfo, history []models.DeploymentHistory) (*models.ManifestResult, error)
}

// NewService creates a new AI service with the given configuration.
func NewService(cfg config.AIConfig) (Service, error) {
	// TODO: Initialize actual LLM client based on provider (openai, claude, azure-openai)
	return &stubService{
		provider: cfg.Provider,
		model:    cfg.Model,
	}, nil
}

type stubService struct {
	provider string
	model    string
}

func (s *stubService) GenerateManifest(ctx context.Context, info ContainerInfo, history []models.DeploymentHistory) (*models.ManifestResult, error) {
	// TODO: Implement actual AI manifest generation
	// 1. Build prompt with system prompt + few-shot examples + container info
	// 2. Call LLM API (OpenAI / Claude)
	// 3. Parse response and extract YAML manifests
	// 4. Validate YAML syntax and security policies
	return nil, fmt.Errorf("AI manifest generation not implemented (provider: %s, model: %s)", s.provider, s.model)
}
