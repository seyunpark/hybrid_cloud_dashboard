package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"

	_ "modernc.org/sqlite"
)

// Store defines the interface for deployment history and settings persistence.
type Store interface {
	Init() error
	Close() error
	SaveDeployment(ctx context.Context, deployment *models.DeploymentHistory) error
	GetDeployHistory(ctx context.Context, limit int) ([]models.DeploymentHistory, error)
	FindSimilar(ctx context.Context, imageName, serviceType string, limit int) ([]models.DeploymentHistory, error)

	// Deployment lifecycle
	GetDeployment(ctx context.Context, id string) (*models.DeploymentHistory, error)
	UpdateDeploymentStatus(ctx context.Context, id string, status string, deletedAt *time.Time) error
	DeleteDeploymentRecord(ctx context.Context, id string) error

	// Settings persistence (key-value)
	SaveSetting(ctx context.Context, key, value string) error
	GetSetting(ctx context.Context, key string) (string, error)
	GetAllSettings(ctx context.Context, prefix string) (map[string]string, error)
	DeleteSetting(ctx context.Context, key string) error

	// Registered clusters persistence
	SaveRegisteredCluster(ctx context.Context, cluster *models.RegisteredCluster) error
	DeleteRegisteredCluster(ctx context.Context, name string) error
	GetRegisteredClusters(ctx context.Context) ([]models.RegisteredCluster, error)

	// Unified deploy history (paginated)
	ListUnifiedHistory(ctx context.Context, offset, limit int) ([]models.UnifiedDeployItem, int, error)

	// Stack deploy persistence
	SaveStackDeploy(ctx context.Context, record *models.StackDeployRecord) error
	GetStackDeploy(ctx context.Context, deployID string) (*models.StackDeployRecord, error)
	UpdateStackDeploy(ctx context.Context, record *models.StackDeployRecord) error
	ListStackDeploys(ctx context.Context, limit int) ([]models.StackDeployRecord, error)
	DeleteStackDeploy(ctx context.Context, deployID string) error

	// Cleanup
	CleanupOldRecords(ctx context.Context, retentionDays int) (int64, error)
}

// NewStore creates a new data store based on the database configuration.
func NewStore(cfg config.DatabaseConfig) (Store, error) {
	switch cfg.Type {
	case "sqlite", "":
		return newSQLiteStore(cfg.Path)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}

type sqliteStore struct {
	path      string
	db        *sql.DB
	stopCh    chan struct{}
	closeOnce sync.Once
}

func newSQLiteStore(path string) (*sqliteStore, error) {
	return &sqliteStore{path: path}, nil
}

func (s *sqliteStore) Init() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating database directory: %w", err)
	}

	db, err := sql.Open("sqlite", s.path)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return fmt.Errorf("setting journal mode: %w", err)
	}

	if err := s.migrate(db); err != nil {
		db.Close()
		return fmt.Errorf("running migrations: %w", err)
	}

	s.db = db

	// Run initial cleanup and start periodic cleanup (every 24h, 30 day retention)
	s.stopCh = make(chan struct{})
	go s.cleanupLoop(30, 24*time.Hour)

	return nil
}

func (s *sqliteStore) migrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS deployment_history (
		id TEXT PRIMARY KEY,
		service_name TEXT NOT NULL DEFAULT '',
		image_name TEXT NOT NULL DEFAULT '',
		image_tag TEXT NOT NULL DEFAULT '',
		service_type TEXT NOT NULL DEFAULT '',
		language TEXT NOT NULL DEFAULT '',
		cpu_request TEXT NOT NULL DEFAULT '',
		cpu_limit TEXT NOT NULL DEFAULT '',
		memory_request TEXT NOT NULL DEFAULT '',
		memory_limit TEXT NOT NULL DEFAULT '',
		replicas INTEGER NOT NULL DEFAULT 1,
		actual_cpu TEXT NOT NULL DEFAULT '',
		actual_memory TEXT NOT NULL DEFAULT '',
		target_cluster TEXT NOT NULL DEFAULT '',
		namespace TEXT NOT NULL DEFAULT 'default',
		deployed_at DATETIME NOT NULL,
		success BOOLEAN NOT NULL DEFAULT 0,
		oom_events INTEGER NOT NULL DEFAULT 0,
		throttle_events INTEGER NOT NULL DEFAULT 0,
		ai_generated BOOLEAN NOT NULL DEFAULT 0,
		ai_confidence REAL NOT NULL DEFAULT 0.0
	);

	CREATE INDEX IF NOT EXISTS idx_deployment_history_image ON deployment_history(image_name);
	CREATE INDEX IF NOT EXISTS idx_deployment_history_service_type ON deployment_history(service_type);
	CREATE INDEX IF NOT EXISTS idx_deployment_history_deployed_at ON deployment_history(deployed_at);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL DEFAULT '',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS registered_clusters (
		name TEXT PRIMARY KEY,
		type TEXT NOT NULL DEFAULT 'kubernetes',
		kubeconfig TEXT NOT NULL DEFAULT '~/.kube/config',
		context TEXT NOT NULL DEFAULT '',
		registry TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS stack_deploys (
		deploy_id       TEXT PRIMARY KEY,
		stack_name      TEXT NOT NULL,
		cluster_name    TEXT NOT NULL DEFAULT '',
		namespace       TEXT NOT NULL DEFAULT 'default',
		container_ids   TEXT NOT NULL DEFAULT '[]',
		create_namespace INTEGER NOT NULL DEFAULT 0,
		prompt          TEXT NOT NULL DEFAULT '',
		status          TEXT NOT NULL DEFAULT 'generating',
		started_at      DATETIME,
		completed_at    DATETIME,
		topology_json   TEXT NOT NULL DEFAULT '',
		manifests_json  TEXT NOT NULL DEFAULT '',
		reasoning       TEXT NOT NULL DEFAULT '',
		confidence      REAL NOT NULL DEFAULT 0.0,
		deploy_order    TEXT NOT NULL DEFAULT '[]',
		services_json   TEXT NOT NULL DEFAULT '{}',
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_stack_deploys_status ON stack_deploys(status);
	CREATE INDEX IF NOT EXISTS idx_stack_deploys_created_at ON stack_deploys(created_at);
	`
	if _, err := db.Exec(query); err != nil {
		return err
	}

	// Add lifecycle columns (idempotent)
	alterStmts := []string{
		`ALTER TABLE deployment_history ADD COLUMN status TEXT NOT NULL DEFAULT 'deployed'`,
		`ALTER TABLE deployment_history ADD COLUMN manifest_json TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE deployment_history ADD COLUMN deleted_at DATETIME`,
	}
	for _, stmt := range alterStmts {
		if _, err := db.Exec(stmt); err != nil && !strings.Contains(err.Error(), "duplicate column") {
			// ignore duplicate column errors for idempotency
		}
	}

	// Backfill status from success bool for existing records
	db.Exec(`UPDATE deployment_history SET status = 'deployed' WHERE success = 1 AND status = ''`)
	db.Exec(`UPDATE deployment_history SET status = 'failed' WHERE success = 0 AND status = ''`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_deployment_history_status ON deployment_history(status)`)

	// Add create_namespace and prompt columns to stack_deploys (idempotent)
	stackAlterStmts := []string{
		`ALTER TABLE stack_deploys ADD COLUMN create_namespace INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE stack_deploys ADD COLUMN prompt TEXT NOT NULL DEFAULT ''`,
	}
	for _, stmt := range stackAlterStmts {
		if _, err := db.Exec(stmt); err != nil && !strings.Contains(err.Error(), "duplicate column") {
			// ignore duplicate column errors for idempotency
		}
	}

	// Normalize timestamps to RFC3339 for consistent cross-table sorting.
	// Go's time.Time default format includes timezone name and monotonic clock
	// (e.g. "2026-03-04 17:12:33 +0900 KST m=+249") which breaks text-based ORDER BY.
	s.normalizeTimestamps(db)

	return nil
}

// normalizeTimestamps rewrites non-RFC3339 timestamps to a sortable UTC format.
func (s *sqliteStore) normalizeTimestamps(db *sql.DB) {
	normalize := func(table, column string) {
		rows, err := db.Query(fmt.Sprintf(
			`SELECT rowid, %s FROM %s WHERE %s NOT LIKE '____-__-__T%%'`,
			column, table, column))
		if err != nil {
			return
		}
		defer rows.Close()

		type row struct {
			rowid int64
			val   string
		}
		var toFix []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.rowid, &r.val); err == nil && r.val != "" {
				toFix = append(toFix, r)
			}
		}
		rows.Close()

		for _, r := range toFix {
			// Try common Go time formats
			var t time.Time
			for _, layout := range []string{
				"2006-01-02 15:04:05.999999999 -0700 MST",
				"2006-01-02 15:04:05.999999999 +0000 UTC",
				"2006-01-02 15:04:05 -0700 MST",
				"2006-01-02 15:04:05",
				time.RFC3339Nano,
				time.RFC3339,
			} {
				// Strip monotonic clock suffix (m=+...)
				raw := r.val
				if idx := strings.Index(raw, " m="); idx > 0 {
					raw = raw[:idx]
				}
				if parsed, err := time.Parse(layout, raw); err == nil {
					t = parsed
					break
				}
			}
			if !t.IsZero() {
				db.Exec(fmt.Sprintf(`UPDATE %s SET %s = ? WHERE rowid = ?`, table, column),
					t.UTC().Format(time.RFC3339Nano), r.rowid)
			}
		}
	}

	normalize("deployment_history", "deployed_at")
	normalize("stack_deploys", "created_at")
	normalize("stack_deploys", "updated_at")
}

func (s *sqliteStore) Close() error {
	var dbErr error
	s.closeOnce.Do(func() {
		if s.stopCh != nil {
			close(s.stopCh)
		}
		if s.db != nil {
			dbErr = s.db.Close()
		}
	})
	return dbErr
}

func (s *sqliteStore) SaveDeployment(ctx context.Context, deployment *models.DeploymentHistory) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Auto-set status from success if not explicitly set
	if deployment.Status == "" {
		if deployment.Success {
			deployment.Status = "deployed"
		} else {
			deployment.Status = "failed"
		}
	}

	query := `
	INSERT INTO deployment_history (
		id, service_name, image_name, image_tag, service_type, language,
		cpu_request, cpu_limit, memory_request, memory_limit,
		replicas, actual_cpu, actual_memory,
		target_cluster, namespace, deployed_at, success,
		status, manifest_json, deleted_at,
		oom_events, throttle_events, ai_generated, ai_confidence
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Normalize timestamp to RFC3339 UTC for consistent sorting
	deployedAtStr := deployment.DeployedAt.UTC().Format(time.RFC3339Nano)

	_, err := s.db.ExecContext(ctx, query,
		deployment.ID, deployment.ServiceName, deployment.ImageName, deployment.ImageTag,
		deployment.ServiceType, deployment.Language,
		deployment.CPURequest, deployment.CPULimit, deployment.MemoryRequest, deployment.MemoryLimit,
		deployment.Replicas, deployment.ActualCPU, deployment.ActualMemory,
		deployment.TargetCluster, deployment.Namespace, deployedAtStr, deployment.Success,
		deployment.Status, deployment.ManifestJSON, deployment.DeletedAt,
		deployment.OOMEvents, deployment.ThrottleEvents, deployment.AIGenerated, deployment.AIConfidence,
	)
	return err
}

func (s *sqliteStore) GetDeployHistory(ctx context.Context, limit int) ([]models.DeploymentHistory, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	if limit <= 0 {
		limit = 50
	}

	query := `SELECT id, service_name, image_name, image_tag, service_type, language,
		cpu_request, cpu_limit, memory_request, memory_limit,
		replicas, actual_cpu, actual_memory,
		target_cluster, namespace, deployed_at, success,
		status, manifest_json, deleted_at,
		oom_events, throttle_events, ai_generated, ai_confidence
		FROM deployment_history ORDER BY deployed_at DESC LIMIT ?`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDeployments(rows)
}

func (s *sqliteStore) FindSimilar(ctx context.Context, imageName, serviceType string, limit int) ([]models.DeploymentHistory, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	if limit <= 0 {
		limit = 5
	}

	query := `SELECT id, service_name, image_name, image_tag, service_type, language,
		cpu_request, cpu_limit, memory_request, memory_limit,
		replicas, actual_cpu, actual_memory,
		target_cluster, namespace, deployed_at, success,
		status, manifest_json, deleted_at,
		oom_events, throttle_events, ai_generated, ai_confidence
		FROM deployment_history
		WHERE success = 1
		  AND (image_name LIKE ? OR service_type = ?)
		ORDER BY deployed_at DESC LIMIT ?`

	likePattern := "%" + imageName + "%"
	rows, err := s.db.QueryContext(ctx, query, likePattern, serviceType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDeployments(rows)
}

// --- Settings persistence ---

func (s *sqliteStore) SaveSetting(ctx context.Context, key, value string) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	query := `INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`
	_, err := s.db.ExecContext(ctx, query, key, value, time.Now().UTC())
	return err
}

func (s *sqliteStore) GetSetting(ctx context.Context, key string) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("database not initialized")
	}
	var value string
	err := s.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (s *sqliteStore) GetAllSettings(ctx context.Context, prefix string) (map[string]string, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	query := "SELECT key, value FROM settings WHERE key LIKE ?"
	rows, err := s.db.QueryContext(ctx, query, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		result[k] = v
	}
	return result, rows.Err()
}

func (s *sqliteStore) DeleteSetting(ctx context.Context, key string) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := s.db.ExecContext(ctx, "DELETE FROM settings WHERE key = ?", key)
	return err
}

// --- Registered clusters persistence ---

func (s *sqliteStore) SaveRegisteredCluster(ctx context.Context, cluster *models.RegisteredCluster) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	query := `INSERT INTO registered_clusters (name, type, kubeconfig, context, registry, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET type = excluded.type, kubeconfig = excluded.kubeconfig,
		context = excluded.context, registry = excluded.registry`
	_, err := s.db.ExecContext(ctx, query,
		cluster.Name, cluster.Type, cluster.Kubeconfig, cluster.Context, cluster.Registry, time.Now().UTC())
	return err
}

func (s *sqliteStore) DeleteRegisteredCluster(ctx context.Context, name string) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := s.db.ExecContext(ctx, "DELETE FROM registered_clusters WHERE name = ?", name)
	return err
}

func (s *sqliteStore) GetRegisteredClusters(ctx context.Context) ([]models.RegisteredCluster, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	rows, err := s.db.QueryContext(ctx, "SELECT name, type, kubeconfig, context, registry, created_at FROM registered_clusters ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.RegisteredCluster
	for rows.Next() {
		var c models.RegisteredCluster
		var createdAt string
		if err := rows.Scan(&c.Name, &c.Type, &c.Kubeconfig, &c.Context, &c.Registry, &createdAt); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		result = append(result, c)
	}
	if result == nil {
		result = []models.RegisteredCluster{}
	}
	return result, rows.Err()
}

// New lifecycle methods

func (s *sqliteStore) GetDeployment(ctx context.Context, id string) (*models.DeploymentHistory, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT id, service_name, image_name, image_tag, service_type, language,
		cpu_request, cpu_limit, memory_request, memory_limit,
		replicas, actual_cpu, actual_memory,
		target_cluster, namespace, deployed_at, success,
		status, manifest_json, deleted_at,
		oom_events, throttle_events, ai_generated, ai_confidence
		FROM deployment_history WHERE id = ?`

	var d models.DeploymentHistory
	var deployedAt string
	var deletedAt sql.NullString
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.ServiceName, &d.ImageName, &d.ImageTag, &d.ServiceType, &d.Language,
		&d.CPURequest, &d.CPULimit, &d.MemoryRequest, &d.MemoryLimit,
		&d.Replicas, &d.ActualCPU, &d.ActualMemory,
		&d.TargetCluster, &d.Namespace, &deployedAt, &d.Success,
		&d.Status, &d.ManifestJSON, &deletedAt,
		&d.OOMEvents, &d.ThrottleEvents, &d.AIGenerated, &d.AIConfidence,
	)
	if err != nil {
		return nil, err
	}
	d.DeployedAt, _ = time.Parse(time.RFC3339, deployedAt)
	if deletedAt.Valid {
		t, _ := time.Parse(time.RFC3339, deletedAt.String)
		d.DeletedAt = &t
	}
	return &d, nil
}

func (s *sqliteStore) UpdateDeploymentStatus(ctx context.Context, id string, status string, deletedAt *time.Time) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE deployment_history SET status = ?, deleted_at = ? WHERE id = ?`,
		status, deletedAt, id)
	return err
}

func (s *sqliteStore) DeleteDeploymentRecord(ctx context.Context, id string) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM deployment_history WHERE id = ?`, id)
	return err
}

// --- Stack deploy persistence ---

func (s *sqliteStore) SaveStackDeploy(ctx context.Context, record *models.StackDeployRecord) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	containerIDsJSON, _ := json.Marshal(record.ContainerIDs)
	deployOrderJSON, _ := json.Marshal(record.DeployOrder)

	query := `INSERT INTO stack_deploys (
		deploy_id, stack_name, cluster_name, namespace, container_ids,
		create_namespace, prompt,
		status, started_at, completed_at,
		topology_json, manifests_json, reasoning, confidence,
		deploy_order, services_json, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	nowStr := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, query,
		record.DeployID, record.StackName, record.ClusterName, record.Namespace, string(containerIDsJSON),
		record.CreateNamespace, record.Prompt,
		record.Status, record.StartedAt, record.CompletedAt,
		record.TopologyJSON, record.ManifestsJSON, record.Reasoning, record.Confidence,
		string(deployOrderJSON), record.ServicesJSON, nowStr, nowStr,
	)
	return err
}

func (s *sqliteStore) GetStackDeploy(ctx context.Context, deployID string) (*models.StackDeployRecord, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	query := `SELECT deploy_id, stack_name, cluster_name, namespace, container_ids,
		create_namespace, prompt,
		status, started_at, completed_at,
		topology_json, manifests_json, reasoning, confidence,
		deploy_order, services_json, created_at, updated_at
		FROM stack_deploys WHERE deploy_id = ?`

	var r models.StackDeployRecord
	var containerIDsJSON, deployOrderJSON string
	var startedAt, completedAt, createdAt, updatedAt sql.NullString

	err := s.db.QueryRowContext(ctx, query, deployID).Scan(
		&r.DeployID, &r.StackName, &r.ClusterName, &r.Namespace, &containerIDsJSON,
		&r.CreateNamespace, &r.Prompt,
		&r.Status, &startedAt, &completedAt,
		&r.TopologyJSON, &r.ManifestsJSON, &r.Reasoning, &r.Confidence,
		&deployOrderJSON, &r.ServicesJSON, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(containerIDsJSON), &r.ContainerIDs)
	json.Unmarshal([]byte(deployOrderJSON), &r.DeployOrder)
	if startedAt.Valid {
		t, _ := time.Parse(time.RFC3339, startedAt.String)
		r.StartedAt = &t
	}
	if completedAt.Valid {
		t, _ := time.Parse(time.RFC3339, completedAt.String)
		r.CompletedAt = &t
	}
	if createdAt.Valid {
		r.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	}
	if updatedAt.Valid {
		r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt.String)
	}
	if r.ContainerIDs == nil {
		r.ContainerIDs = []string{}
	}
	if r.DeployOrder == nil {
		r.DeployOrder = []string{}
	}
	return &r, nil
}

func (s *sqliteStore) UpdateStackDeploy(ctx context.Context, record *models.StackDeployRecord) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	deployOrderJSON, _ := json.Marshal(record.DeployOrder)

	query := `UPDATE stack_deploys SET
		status = ?, completed_at = ?,
		topology_json = ?, manifests_json = ?, reasoning = ?, confidence = ?,
		deploy_order = ?, services_json = ?, updated_at = ?
		WHERE deploy_id = ?`

	_, err := s.db.ExecContext(ctx, query,
		record.Status, record.CompletedAt,
		record.TopologyJSON, record.ManifestsJSON, record.Reasoning, record.Confidence,
		string(deployOrderJSON), record.ServicesJSON, time.Now().UTC().Format(time.RFC3339Nano),
		record.DeployID,
	)
	return err
}

func (s *sqliteStore) ListStackDeploys(ctx context.Context, limit int) ([]models.StackDeployRecord, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	if limit <= 0 {
		limit = 100
	}

	query := `SELECT deploy_id, stack_name, cluster_name, namespace, container_ids,
		create_namespace, prompt,
		status, started_at, completed_at,
		topology_json, manifests_json, reasoning, confidence,
		deploy_order, services_json, created_at, updated_at
		FROM stack_deploys ORDER BY created_at DESC LIMIT ?`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.StackDeployRecord
	for rows.Next() {
		var r models.StackDeployRecord
		var containerIDsJSON, deployOrderJSON string
		var startedAt, completedAt, createdAt, updatedAt sql.NullString

		if err := rows.Scan(
			&r.DeployID, &r.StackName, &r.ClusterName, &r.Namespace, &containerIDsJSON,
			&r.CreateNamespace, &r.Prompt,
			&r.Status, &startedAt, &completedAt,
			&r.TopologyJSON, &r.ManifestsJSON, &r.Reasoning, &r.Confidence,
			&deployOrderJSON, &r.ServicesJSON, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(containerIDsJSON), &r.ContainerIDs)
		json.Unmarshal([]byte(deployOrderJSON), &r.DeployOrder)
		if startedAt.Valid {
			t, _ := time.Parse(time.RFC3339, startedAt.String)
			r.StartedAt = &t
		}
		if completedAt.Valid {
			t, _ := time.Parse(time.RFC3339, completedAt.String)
			r.CompletedAt = &t
		}
		if createdAt.Valid {
			r.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
		}
		if updatedAt.Valid {
			r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt.String)
		}
		if r.ContainerIDs == nil {
			r.ContainerIDs = []string{}
		}
		if r.DeployOrder == nil {
			r.DeployOrder = []string{}
		}
		results = append(results, r)
	}
	if results == nil {
		results = []models.StackDeployRecord{}
	}
	return results, rows.Err()
}

func (s *sqliteStore) DeleteStackDeploy(ctx context.Context, deployID string) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM stack_deploys WHERE deploy_id = ?`, deployID)
	return err
}

// ListUnifiedHistory returns a paginated, chronologically-ordered list of both
// single and stack deploys merged via UNION ALL.
func (s *sqliteStore) ListUnifiedHistory(ctx context.Context, offset, limit int) ([]models.UnifiedDeployItem, int, error) {
	if s.db == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Exclude deployment_history records that were created by stack deploys
	// (their ID pattern is "{stack_deploy_id}_{service_name}")
	stackFilter := `NOT EXISTS (
		SELECT 1 FROM stack_deploys sd
		WHERE deployment_history.id LIKE sd.deploy_id || '_%'
	)`

	// Total count
	var total int
	countQuery := fmt.Sprintf(`SELECT
		(SELECT COUNT(*) FROM stack_deploys) +
		(SELECT COUNT(*) FROM deployment_history WHERE %s) AS total`, stackFilter)
	if err := s.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting unified history: %w", err)
	}

	// Paginated union query
	query := fmt.Sprintf(`
	SELECT id, type, name, cluster, namespace, status,
	       ai_generated, confidence, deployed_at,
	       deploy_order_json,
	       image_name, image_tag, replicas
	FROM (
		SELECT
			deploy_id AS id,
			'stack' AS type,
			stack_name AS name,
			cluster_name AS cluster,
			namespace,
			status,
			1 AS ai_generated,
			confidence,
			created_at AS deployed_at,
			deploy_order AS deploy_order_json,
			'' AS image_name,
			'' AS image_tag,
			0 AS replicas
		FROM stack_deploys
		UNION ALL
		SELECT
			id,
			'single' AS type,
			service_name AS name,
			target_cluster AS cluster,
			namespace,
			status,
			ai_generated,
			ai_confidence AS confidence,
			deployed_at,
			'' AS deploy_order_json,
			image_name,
			image_tag,
			replicas
		FROM deployment_history
		WHERE %s
	)
	ORDER BY deployed_at DESC
	LIMIT ? OFFSET ?`, stackFilter)

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("querying unified history: %w", err)
	}
	defer rows.Close()

	var results []models.UnifiedDeployItem
	for rows.Next() {
		var item models.UnifiedDeployItem
		var deployedAt string
		var deployOrderJSON string
		var imageName, imageTag string
		var replicas int
		var aiGenerated bool

		if err := rows.Scan(
			&item.ID, &item.Type, &item.Name, &item.Cluster, &item.Namespace, &item.Status,
			&aiGenerated, &item.Confidence, &deployedAt,
			&deployOrderJSON,
			&imageName, &imageTag, &replicas,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning unified row: %w", err)
		}

		item.AIGenerated = aiGenerated
		item.DeployedAt, _ = time.Parse(time.RFC3339, deployedAt)

		if item.Type == "stack" {
			var deployOrder []string
			json.Unmarshal([]byte(deployOrderJSON), &deployOrder)
			// Filter out internal entries like "_namespace"
			filtered := make([]string, 0, len(deployOrder))
			for _, name := range deployOrder {
				if !strings.HasPrefix(name, "_") {
					filtered = append(filtered, name)
				}
			}
			item.StackDetail = &models.StackDeployBrief{
				ServiceCount: len(filtered),
				Services:     filtered,
				DeployOrder:  filtered,
			}
			item.ImageSummary = fmt.Sprintf("%d services", len(filtered))
		} else {
			item.SingleDetail = &models.SingleDeployBrief{
				ImageName: imageName,
				ImageTag:  imageTag,
				Replicas:  replicas,
			}
			if imageTag != "" {
				item.ImageSummary = imageName + ":" + imageTag
			} else {
				item.ImageSummary = imageName
			}
		}

		results = append(results, item)
	}

	if results == nil {
		results = []models.UnifiedDeployItem{}
	}

	return results, total, rows.Err()
}

// CleanupOldRecords deletes deployment history and stack deploy records
// older than the given retention period.
func (s *sqliteStore) CleanupOldRecords(ctx context.Context, retentionDays int) (int64, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays).Format(time.RFC3339Nano)
	var totalDeleted int64

	// Delete old deployment_history records
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM deployment_history WHERE deployed_at < ?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("cleaning deployment_history: %w", err)
	}
	n, _ := res.RowsAffected()
	totalDeleted += n

	// Delete old stack_deploys records (only completed/failed/deleted/undeployed)
	res, err = s.db.ExecContext(ctx,
		`DELETE FROM stack_deploys WHERE created_at < ? AND status IN ('deleted', 'failed', 'undeployed')`,
		cutoff)
	if err != nil {
		return totalDeleted, fmt.Errorf("cleaning stack_deploys: %w", err)
	}
	n, _ = res.RowsAffected()
	totalDeleted += n

	return totalDeleted, nil
}

func (s *sqliteStore) cleanupLoop(retentionDays int, interval time.Duration) {
	// Run once at startup
	ctx := context.Background()
	if deleted, err := s.CleanupOldRecords(ctx, retentionDays); err != nil {
		slog.Error("initial history cleanup failed", "error", err)
	} else if deleted > 0 {
		slog.Info("cleaned up old history records", "deleted", deleted, "retention_days", retentionDays)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if deleted, err := s.CleanupOldRecords(ctx, retentionDays); err != nil {
				slog.Error("periodic history cleanup failed", "error", err)
			} else if deleted > 0 {
				slog.Info("cleaned up old history records", "deleted", deleted, "retention_days", retentionDays)
			}
		}
	}
}

func scanDeployments(rows *sql.Rows) ([]models.DeploymentHistory, error) {
	var results []models.DeploymentHistory
	for rows.Next() {
		var d models.DeploymentHistory
		var deployedAt string
		var deletedAt sql.NullString
		err := rows.Scan(
			&d.ID, &d.ServiceName, &d.ImageName, &d.ImageTag, &d.ServiceType, &d.Language,
			&d.CPURequest, &d.CPULimit, &d.MemoryRequest, &d.MemoryLimit,
			&d.Replicas, &d.ActualCPU, &d.ActualMemory,
			&d.TargetCluster, &d.Namespace, &deployedAt, &d.Success,
			&d.Status, &d.ManifestJSON, &deletedAt,
			&d.OOMEvents, &d.ThrottleEvents, &d.AIGenerated, &d.AIConfidence,
		)
		if err != nil {
			return nil, err
		}
		d.DeployedAt, _ = time.Parse(time.RFC3339, deployedAt)
		if deletedAt.Valid {
			t, _ := time.Parse(time.RFC3339, deletedAt.String)
			d.DeletedAt = &t
		}
		results = append(results, d)
	}
	if results == nil {
		results = []models.DeploymentHistory{}
	}
	return results, rows.Err()
}
