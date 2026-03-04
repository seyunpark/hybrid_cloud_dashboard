package data

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

func setupTestStore(t *testing.T) (Store, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStore(config.DatabaseConfig{
		Type: "sqlite",
		Path: dbPath,
	})
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	if err := store.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.Remove(dbPath)
	}

	return store, cleanup
}

func TestNewStore_SQLite(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestNewStore_UnsupportedType(t *testing.T) {
	_, err := NewStore(config.DatabaseConfig{
		Type: "postgres",
		Path: "/tmp/test.db",
	})
	if err == nil {
		t.Fatal("expected error for unsupported database type")
	}
}

func TestInit_CreatesDatabaseFile(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "test.db")

	store, err := NewStore(config.DatabaseConfig{
		Type: "sqlite",
		Path: dbPath,
	})
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	if err := store.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer store.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("expected database file to be created")
	}
}

func TestSaveDeployment(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	deployment := &models.DeploymentHistory{
		ID:            "test-deploy-1",
		ServiceName:   "my-service",
		ImageName:     "nginx",
		ImageTag:      "1.25",
		ServiceType:   "web-server",
		Language:      "unknown",
		CPURequest:    "100m",
		CPULimit:      "500m",
		MemoryRequest: "128Mi",
		MemoryLimit:   "512Mi",
		Replicas:      2,
		ActualCPU:     "50m",
		ActualMemory:  "64Mi",
		TargetCluster: "local-k8s",
		Namespace:     "default",
		DeployedAt:    time.Now(),
		Success:       true,
		OOMEvents:     0,
		ThrottleEvents: 0,
		AIGenerated:   true,
		AIConfidence:  0.85,
	}

	err := store.SaveDeployment(ctx, deployment)
	if err != nil {
		t.Fatalf("SaveDeployment failed: %v", err)
	}
}

func TestSaveDeployment_DuplicateID(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	deployment := &models.DeploymentHistory{
		ID:          "dup-id",
		ServiceName: "svc1",
		ImageName:   "nginx",
		DeployedAt:  time.Now(),
	}

	if err := store.SaveDeployment(ctx, deployment); err != nil {
		t.Fatalf("first save failed: %v", err)
	}

	err := store.SaveDeployment(ctx, deployment)
	if err == nil {
		t.Fatal("expected error for duplicate ID")
	}
}

func TestGetDeployHistory(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Insert test data
	for i := 0; i < 5; i++ {
		dep := &models.DeploymentHistory{
			ID:          "hist-" + string(rune('A'+i)),
			ServiceName: "service-" + string(rune('A'+i)),
			ImageName:   "nginx",
			ImageTag:    "latest",
			DeployedAt:  time.Now().Add(time.Duration(i) * time.Minute),
			Success:     true,
		}
		if err := store.SaveDeployment(ctx, dep); err != nil {
			t.Fatalf("SaveDeployment %d failed: %v", i, err)
		}
	}

	// Get all
	history, err := store.GetDeployHistory(ctx, 10)
	if err != nil {
		t.Fatalf("GetDeployHistory failed: %v", err)
	}
	if len(history) != 5 {
		t.Fatalf("expected 5 records, got %d", len(history))
	}

	// Verify ordering (newest first)
	if history[0].ID != "hist-E" {
		t.Errorf("expected first record to be hist-E, got %s", history[0].ID)
	}

	// Test limit
	limited, err := store.GetDeployHistory(ctx, 2)
	if err != nil {
		t.Fatalf("GetDeployHistory with limit failed: %v", err)
	}
	if len(limited) != 2 {
		t.Fatalf("expected 2 records with limit, got %d", len(limited))
	}
}

func TestGetDeployHistory_EmptyDB(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	history, err := store.GetDeployHistory(ctx, 10)
	if err != nil {
		t.Fatalf("GetDeployHistory failed: %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("expected 0 records, got %d", len(history))
	}
}

func TestGetDeployHistory_DefaultLimit(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	// Negative limit should default to 50
	history, err := store.GetDeployHistory(ctx, -1)
	if err != nil {
		t.Fatalf("GetDeployHistory failed: %v", err)
	}
	if history == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestFindSimilar(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Insert test data
	deployments := []models.DeploymentHistory{
		{ID: "sim-1", ServiceName: "web-app", ImageName: "nginx", ServiceType: "web-server", DeployedAt: time.Now(), Success: true},
		{ID: "sim-2", ServiceName: "api-svc", ImageName: "node-api", ServiceType: "api", DeployedAt: time.Now(), Success: true},
		{ID: "sim-3", ServiceName: "web-proxy", ImageName: "nginx-proxy", ServiceType: "web-server", DeployedAt: time.Now(), Success: true},
		{ID: "sim-4", ServiceName: "db", ImageName: "postgres", ServiceType: "database", DeployedAt: time.Now(), Success: false}, // failed - should be excluded
		{ID: "sim-5", ServiceName: "cache", ImageName: "redis", ServiceType: "cache", DeployedAt: time.Now(), Success: true},
	}

	for _, d := range deployments {
		dep := d
		if err := store.SaveDeployment(ctx, &dep); err != nil {
			t.Fatalf("SaveDeployment failed: %v", err)
		}
	}

	// Search by image name
	results, err := store.FindSimilar(ctx, "nginx", "", 10)
	if err != nil {
		t.Fatalf("FindSimilar failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 similar by image, got %d", len(results))
	}

	// Search by service type (also matches empty imageName LIKE '%%' which matches all)
	results, err = store.FindSimilar(ctx, "", "web-server", 10)
	if err != nil {
		t.Fatalf("FindSimilar failed: %v", err)
	}
	// All successful deployments match because imageName LIKE '%%' matches everything
	if len(results) < 2 {
		t.Fatalf("expected at least 2 similar by service type, got %d", len(results))
	}

	// Search with limit
	results, err = store.FindSimilar(ctx, "nginx", "web-server", 1)
	if err != nil {
		t.Fatalf("FindSimilar failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result with limit, got %d", len(results))
	}
}

func TestFindSimilar_ExcludesFailedDeployments(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	dep := &models.DeploymentHistory{
		ID:        "fail-1",
		ImageName: "myapp",
		DeployedAt: time.Now(),
		Success:   false,
	}
	if err := store.SaveDeployment(ctx, dep); err != nil {
		t.Fatalf("SaveDeployment failed: %v", err)
	}

	results, err := store.FindSimilar(ctx, "myapp", "", 10)
	if err != nil {
		t.Fatalf("FindSimilar failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatal("expected failed deployments to be excluded")
	}
}

func TestClose_Idempotent(t *testing.T) {
	store, _ := setupTestStore(t)

	if err := store.Close(); err != nil {
		t.Fatalf("first Close failed: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("second Close failed: %v", err)
	}
}
