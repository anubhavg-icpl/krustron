// Package gitops provides GitOps application management
// Author: Anubhav Gain <anubhavg@infopercept.com>
package gitops

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/anubhavg-icpl/krustron/pkg/config"
	"github.com/anubhavg-icpl/krustron/pkg/database"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/anubhavg-icpl/krustron/pkg/kube"
	"github.com/anubhavg-icpl/krustron/pkg/logger"
	"go.uber.org/zap"
)

// Service provides GitOps management functionality
type Service struct {
	db          *database.PostgresDB
	kubeManager *kube.ClientManager
	config      *config.GitOpsConfig
}

// NewService creates a new GitOps service
func NewService(db *database.PostgresDB, kubeManager *kube.ClientManager, cfg *config.GitOpsConfig) *Service {
	return &Service{
		db:          db,
		kubeManager: kubeManager,
		config:      cfg,
	}
}

// Application represents a GitOps application
type Application struct {
	ID           string            `json:"id" db:"id"`
	Name         string            `json:"name" db:"name"`
	DisplayName  string            `json:"display_name" db:"display_name"`
	Description  string            `json:"description" db:"description"`
	ClusterID    string            `json:"cluster_id" db:"cluster_id"`
	Namespace    string            `json:"namespace" db:"namespace"`
	SourceType   string            `json:"source_type" db:"source_type"`
	RepoURL      string            `json:"repo_url" db:"repo_url"`
	RepoBranch   string            `json:"repo_branch" db:"repo_branch"`
	RepoPath     string            `json:"repo_path" db:"repo_path"`
	HelmChart    string            `json:"helm_chart" db:"helm_chart"`
	HelmRepo     string            `json:"helm_repo" db:"helm_repo"`
	HelmVersion  string            `json:"helm_version" db:"helm_version"`
	ValuesYAML   string            `json:"values_yaml,omitempty" db:"values_yaml"`
	SyncPolicy   string            `json:"sync_policy" db:"sync_policy"`
	AutoSync     bool              `json:"auto_sync" db:"auto_sync"`
	Prune        bool              `json:"prune" db:"prune"`
	SelfHeal     bool              `json:"self_heal" db:"self_heal"`
	Status       string            `json:"status" db:"status"`
	HealthStatus string            `json:"health_status" db:"health_status"`
	SyncStatus   string            `json:"sync_status" db:"sync_status"`
	LastSyncAt   *time.Time        `json:"last_sync_at" db:"last_sync_at"`
	Labels       map[string]string `json:"labels" db:"labels"`
	Annotations  map[string]string `json:"annotations" db:"annotations"`
	CreatedBy    string            `json:"created_by" db:"created_by"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at"`
}

// ListFilters contains filters for listing applications
type ListFilters struct {
	Page      int
	Limit     int
	ClusterID string
	Namespace string
	Status    string
}

// CreateRequest contains application creation data
type CreateRequest struct {
	Name        string            `json:"name" binding:"required"`
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	ClusterID   string            `json:"cluster_id" binding:"required"`
	Namespace   string            `json:"namespace" binding:"required"`
	SourceType  string            `json:"source_type" binding:"required,oneof=git helm"`
	RepoURL     string            `json:"repo_url"`
	RepoBranch  string            `json:"repo_branch"`
	RepoPath    string            `json:"repo_path"`
	HelmChart   string            `json:"helm_chart"`
	HelmRepo    string            `json:"helm_repo"`
	HelmVersion string            `json:"helm_version"`
	ValuesYAML  string            `json:"values_yaml"`
	AutoSync    bool              `json:"auto_sync"`
	Prune       bool              `json:"prune"`
	SelfHeal    bool              `json:"self_heal"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedBy   string            `json:"-"`
}

// UpdateRequest contains application update data
type UpdateRequest struct {
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	RepoBranch  string            `json:"repo_branch"`
	RepoPath    string            `json:"repo_path"`
	HelmVersion string            `json:"helm_version"`
	ValuesYAML  string            `json:"values_yaml"`
	AutoSync    *bool             `json:"auto_sync"`
	Prune       *bool             `json:"prune"`
	SelfHeal    *bool             `json:"self_heal"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// SyncRequest contains sync request data
type SyncRequest struct {
	Revision string   `json:"revision"`
	Prune    bool     `json:"prune"`
	DryRun   bool     `json:"dry_run"`
	Resources []string `json:"resources"`
}

// SyncStatus represents sync status
type SyncStatus struct {
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	Revision     string    `json:"revision"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at,omitempty"`
	ResourcesOut []Resource `json:"resources,omitempty"`
}

// Resource represents a Kubernetes resource
type Resource struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Health    string `json:"health"`
	Message   string `json:"message,omitempty"`
}

// List returns all applications with filters
func (s *Service) List(ctx context.Context, filters *ListFilters) ([]Application, int, error) {
	query := `
		SELECT id, name, display_name, description, cluster_id, namespace,
		       source_type, repo_url, repo_branch, repo_path, helm_chart,
		       helm_repo, helm_version, sync_policy, auto_sync, prune,
		       self_heal, status, health_status, sync_status, last_sync_at,
		       labels, annotations, created_by, created_at, updated_at
		FROM applications
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM applications WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if filters.ClusterID != "" {
		argCount++
		query += " AND cluster_id = $" + string(rune('0'+argCount))
		countQuery += " AND cluster_id = $" + string(rune('0'+argCount))
		args = append(args, filters.ClusterID)
	}

	if filters.Namespace != "" {
		argCount++
		query += " AND namespace = $" + string(rune('0'+argCount))
		countQuery += " AND namespace = $" + string(rune('0'+argCount))
		args = append(args, filters.Namespace)
	}

	if filters.Status != "" {
		argCount++
		query += " AND status = $" + string(rune('0'+argCount))
		countQuery += " AND status = $" + string(rune('0'+argCount))
		args = append(args, filters.Status)
	}

	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to count applications")
	}

	offset := (filters.Page - 1) * filters.Limit
	query += " ORDER BY created_at DESC LIMIT $" + string(rune('0'+argCount+1)) + " OFFSET $" + string(rune('0'+argCount+2))
	args = append(args, filters.Limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to query applications")
	}
	defer rows.Close()

	var apps []Application
	for rows.Next() {
		var app Application
		var labels, annotations []byte
		var lastSyncAt sql.NullTime

		if err := rows.Scan(
			&app.ID, &app.Name, &app.DisplayName, &app.Description, &app.ClusterID,
			&app.Namespace, &app.SourceType, &app.RepoURL, &app.RepoBranch,
			&app.RepoPath, &app.HelmChart, &app.HelmRepo, &app.HelmVersion,
			&app.SyncPolicy, &app.AutoSync, &app.Prune, &app.SelfHeal,
			&app.Status, &app.HealthStatus, &app.SyncStatus, &lastSyncAt,
			&labels, &annotations, &app.CreatedBy, &app.CreatedAt, &app.UpdatedAt,
		); err != nil {
			return nil, 0, errors.DatabaseWrap(err, "failed to scan application")
		}

		if lastSyncAt.Valid {
			app.LastSyncAt = &lastSyncAt.Time
		}
		json.Unmarshal(labels, &app.Labels)
		json.Unmarshal(annotations, &app.Annotations)

		apps = append(apps, app)
	}

	return apps, total, nil
}

// Get returns a single application
func (s *Service) Get(ctx context.Context, id string) (*Application, error) {
	query := `
		SELECT id, name, display_name, description, cluster_id, namespace,
		       source_type, repo_url, repo_branch, repo_path, helm_chart,
		       helm_repo, helm_version, values_yaml, sync_policy, auto_sync,
		       prune, self_heal, status, health_status, sync_status, last_sync_at,
		       labels, annotations, created_by, created_at, updated_at
		FROM applications WHERE id = $1
	`

	var app Application
	var labels, annotations []byte
	var lastSyncAt sql.NullTime
	var valuesYAML sql.NullString

	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&app.ID, &app.Name, &app.DisplayName, &app.Description, &app.ClusterID,
		&app.Namespace, &app.SourceType, &app.RepoURL, &app.RepoBranch,
		&app.RepoPath, &app.HelmChart, &app.HelmRepo, &app.HelmVersion,
		&valuesYAML, &app.SyncPolicy, &app.AutoSync, &app.Prune, &app.SelfHeal,
		&app.Status, &app.HealthStatus, &app.SyncStatus, &lastSyncAt,
		&labels, &annotations, &app.CreatedBy, &app.CreatedAt, &app.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("application", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get application")
	}

	if lastSyncAt.Valid {
		app.LastSyncAt = &lastSyncAt.Time
	}
	if valuesYAML.Valid {
		app.ValuesYAML = valuesYAML.String
	}
	json.Unmarshal(labels, &app.Labels)
	json.Unmarshal(annotations, &app.Annotations)

	return &app, nil
}

// Create creates a new application
func (s *Service) Create(ctx context.Context, req *CreateRequest) (*Application, error) {
	labels, _ := json.Marshal(req.Labels)
	annotations, _ := json.Marshal(req.Annotations)

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Name
	}

	repoBranch := req.RepoBranch
	if repoBranch == "" {
		repoBranch = "main"
	}

	repoPath := req.RepoPath
	if repoPath == "" {
		repoPath = "."
	}

	query := `
		INSERT INTO applications (name, display_name, description, cluster_id, namespace,
		                         source_type, repo_url, repo_branch, repo_path,
		                         helm_chart, helm_repo, helm_version, values_yaml,
		                         auto_sync, prune, self_heal, labels, annotations, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, status, health_status, sync_status, created_at, updated_at
	`

	var app Application
	if err := s.db.QueryRowContext(ctx, query,
		req.Name, displayName, req.Description, req.ClusterID, req.Namespace,
		req.SourceType, req.RepoURL, repoBranch, repoPath,
		req.HelmChart, req.HelmRepo, req.HelmVersion, req.ValuesYAML,
		req.AutoSync, req.Prune, req.SelfHeal, labels, annotations, req.CreatedBy,
	).Scan(&app.ID, &app.Status, &app.HealthStatus, &app.SyncStatus, &app.CreatedAt, &app.UpdatedAt); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to create application")
	}

	app.Name = req.Name
	app.DisplayName = displayName
	app.Description = req.Description
	app.ClusterID = req.ClusterID
	app.Namespace = req.Namespace
	app.SourceType = req.SourceType
	app.RepoURL = req.RepoURL
	app.RepoBranch = repoBranch
	app.RepoPath = repoPath
	app.HelmChart = req.HelmChart
	app.HelmRepo = req.HelmRepo
	app.HelmVersion = req.HelmVersion
	app.AutoSync = req.AutoSync
	app.Prune = req.Prune
	app.SelfHeal = req.SelfHeal
	app.Labels = req.Labels
	app.Annotations = req.Annotations
	app.CreatedBy = req.CreatedBy

	logger.Info("Application created",
		zap.String("app_id", app.ID),
		zap.String("name", app.Name),
	)

	return &app, nil
}

// Update updates an application
func (s *Service) Update(ctx context.Context, id string, req *UpdateRequest) (*Application, error) {
	labels, _ := json.Marshal(req.Labels)
	annotations, _ := json.Marshal(req.Annotations)

	query := `
		UPDATE applications
		SET display_name = COALESCE(NULLIF($2, ''), display_name),
		    description = COALESCE(NULLIF($3, ''), description),
		    repo_branch = COALESCE(NULLIF($4, ''), repo_branch),
		    repo_path = COALESCE(NULLIF($5, ''), repo_path),
		    helm_version = COALESCE(NULLIF($6, ''), helm_version),
		    values_yaml = COALESCE(NULLIF($7, ''), values_yaml),
		    auto_sync = COALESCE($8, auto_sync),
		    prune = COALESCE($9, prune),
		    self_heal = COALESCE($10, self_heal),
		    labels = COALESCE($11, labels),
		    annotations = COALESCE($12, annotations),
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.db.ExecContext(ctx, query, id,
		req.DisplayName, req.Description, req.RepoBranch, req.RepoPath,
		req.HelmVersion, req.ValuesYAML, req.AutoSync, req.Prune, req.SelfHeal,
		labels, annotations,
	)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to update application")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.NotFound("application", id)
	}

	return s.Get(ctx, id)
}

// Delete deletes an application
func (s *Service) Delete(ctx context.Context, id string, cascade bool) error {
	query := "DELETE FROM applications WHERE id = $1"
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to delete application")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFound("application", id)
	}

	logger.Info("Application deleted", zap.String("app_id", id))
	return nil
}

// Sync triggers a sync for an application
func (s *Service) Sync(ctx context.Context, id string, req *SyncRequest) (*SyncStatus, error) {
	app, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update sync status
	now := time.Now()
	query := `
		UPDATE applications
		SET sync_status = 'syncing',
		    last_sync_at = $2,
		    updated_at = NOW()
		WHERE id = $1
	`
	s.db.ExecContext(ctx, query, id, now)

	// In a real implementation, this would trigger ArgoCD sync
	status := &SyncStatus{
		Status:     "syncing",
		Message:    "Sync initiated",
		Revision:   req.Revision,
		StartedAt:  now,
	}

	logger.Info("Application sync triggered",
		zap.String("app_id", id),
		zap.String("name", app.Name),
	)

	return status, nil
}

// GetStatus returns application status
func (s *Service) GetStatus(ctx context.Context, id string) (*AppStatus, error) {
	app, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &AppStatus{
		Status:       app.Status,
		HealthStatus: app.HealthStatus,
		SyncStatus:   app.SyncStatus,
		LastSyncAt:   app.LastSyncAt,
	}, nil
}

// AppStatus represents application status
type AppStatus struct {
	Status       string     `json:"status"`
	HealthStatus string     `json:"health_status"`
	SyncStatus   string     `json:"sync_status"`
	LastSyncAt   *time.Time `json:"last_sync_at"`
}

// GetResources returns resources managed by an application
func (s *Service) GetResources(ctx context.Context, id string) ([]Resource, error) {
	// In a real implementation, this would query ArgoCD for resources
	return []Resource{
		{Kind: "Deployment", Name: "app", Namespace: "default", Status: "Synced", Health: "Healthy"},
		{Kind: "Service", Name: "app-svc", Namespace: "default", Status: "Synced", Health: "Healthy"},
		{Kind: "ConfigMap", Name: "app-config", Namespace: "default", Status: "Synced", Health: "Healthy"},
	}, nil
}

// GetEvents returns events for an application
func (s *Service) GetEvents(ctx context.Context, id string) ([]Event, error) {
	// In a real implementation, this would query ArgoCD events
	return []Event{
		{Type: "Sync", Message: "Sync completed", Timestamp: time.Now()},
	}, nil
}

// Event represents an application event
type Event struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// GetManifests returns manifests for an application
func (s *Service) GetManifests(ctx context.Context, id, revision string) ([]Manifest, error) {
	// In a real implementation, this would fetch manifests from git or helm
	return []Manifest{
		{Kind: "Deployment", Name: "app", Content: "apiVersion: apps/v1\nkind: Deployment\n..."},
	}, nil
}

// Manifest represents a Kubernetes manifest
type Manifest struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// GetDiff returns the diff between live and desired state
func (s *Service) GetDiff(ctx context.Context, id string) (*Diff, error) {
	// In a real implementation, this would compute the diff
	return &Diff{
		HasDiff: false,
		Diffs:   []ResourceDiff{},
	}, nil
}

// Diff represents the diff between live and desired state
type Diff struct {
	HasDiff bool           `json:"has_diff"`
	Diffs   []ResourceDiff `json:"diffs"`
}

// ResourceDiff represents a diff for a single resource
type ResourceDiff struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	Diff string `json:"diff"`
}
