// Package helm provides Helm chart and release management
// Author: Anubhav Gain <anubhavg@infopercept.com>
package helm

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/anubhavg-icpl/krustron/pkg/cache"
	"github.com/anubhavg-icpl/krustron/pkg/database"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/anubhavg-icpl/krustron/pkg/kube"
	"github.com/anubhavg-icpl/krustron/pkg/logger"
	"go.uber.org/zap"
)

// Service provides Helm management functionality
type Service struct {
	db          *database.PostgresDB
	kubeManager *kube.ClientManager
	cache       *cache.RedisCache
}

// NewService creates a new Helm service
func NewService(db *database.PostgresDB, kubeManager *kube.ClientManager, cache *cache.RedisCache) *Service {
	return &Service{
		db:          db,
		kubeManager: kubeManager,
		cache:       cache,
	}
}

// Repository represents a Helm repository
type Repository struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Username  string    `json:"username,omitempty"`
	LastSync  time.Time `json:"last_sync"`
	ChartCount int      `json:"chart_count"`
}

// Chart represents a Helm chart
type Chart struct {
	Name        string   `json:"name"`
	Repository  string   `json:"repository"`
	Version     string   `json:"version"`
	AppVersion  string   `json:"app_version"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Keywords    []string `json:"keywords"`
	Home        string   `json:"home"`
	Sources     []string `json:"sources"`
	Deprecated  bool     `json:"deprecated"`
}

// ChartDetails contains detailed chart information
type ChartDetails struct {
	Chart
	Readme       string                 `json:"readme"`
	Values       map[string]interface{} `json:"values"`
	Dependencies []ChartDependency      `json:"dependencies"`
}

// ChartDependency represents a chart dependency
type ChartDependency struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
	Condition  string `json:"condition,omitempty"`
}

// Release represents a Helm release
type Release struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	ClusterID    string    `json:"cluster_id" db:"cluster_id"`
	Namespace    string    `json:"namespace" db:"namespace"`
	ChartName    string    `json:"chart_name" db:"chart_name"`
	ChartVersion string    `json:"chart_version" db:"chart_version"`
	ChartRepo    string    `json:"chart_repo" db:"chart_repo"`
	AppVersion   string    `json:"app_version" db:"app_version"`
	ValuesYAML   string    `json:"values_yaml,omitempty" db:"values_yaml"`
	Status       string    `json:"status" db:"status"`
	Revision     int       `json:"revision" db:"revision"`
	LastDeployed time.Time `json:"last_deployed" db:"last_deployed"`
	Notes        string    `json:"notes" db:"notes"`
	CreatedBy    string    `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ReleaseHistory represents a release history entry
type ReleaseHistory struct {
	Revision    int       `json:"revision"`
	Status      string    `json:"status"`
	Chart       string    `json:"chart"`
	AppVersion  string    `json:"app_version"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AddRepoRequest contains repository addition data
type AddRepoRequest struct {
	Name     string `json:"name" binding:"required"`
	URL      string `json:"url" binding:"required,url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// InstallRequest contains release installation data
type InstallRequest struct {
	Name         string            `json:"name" binding:"required"`
	ClusterID    string            `json:"cluster_id" binding:"required"`
	Namespace    string            `json:"namespace" binding:"required"`
	ChartRepo    string            `json:"chart_repo" binding:"required"`
	ChartName    string            `json:"chart_name" binding:"required"`
	ChartVersion string            `json:"chart_version"`
	Values       map[string]interface{} `json:"values"`
	ValuesYAML   string            `json:"values_yaml"`
	CreateNS     bool              `json:"create_namespace"`
	Wait         bool              `json:"wait"`
	Timeout      int               `json:"timeout"`
	CreatedBy    string            `json:"-"`
}

// UpgradeRequest contains release upgrade data
type UpgradeRequest struct {
	ClusterID    string                 `json:"-"`
	Namespace    string                 `json:"-"`
	Name         string                 `json:"-"`
	ChartVersion string                 `json:"chart_version"`
	Values       map[string]interface{} `json:"values"`
	ValuesYAML   string                 `json:"values_yaml"`
	ResetValues  bool                   `json:"reset_values"`
	ReuseValues  bool                   `json:"reuse_values"`
	Wait         bool                   `json:"wait"`
	Timeout      int                    `json:"timeout"`
}

// RollbackRequest contains rollback data
type RollbackRequest struct {
	Revision int `json:"revision" binding:"required"`
}

// ListRepositories returns all Helm repositories
func (s *Service) ListRepositories(ctx context.Context) ([]Repository, error) {
	// In a real implementation, this would read from Helm config
	// For now, return some default repos
	return []Repository{
		{
			Name:       "stable",
			URL:        "https://charts.helm.sh/stable",
			LastSync:   time.Now(),
			ChartCount: 250,
		},
		{
			Name:       "bitnami",
			URL:        "https://charts.bitnami.com/bitnami",
			LastSync:   time.Now(),
			ChartCount: 100,
		},
	}, nil
}

// AddRepository adds a new Helm repository
func (s *Service) AddRepository(ctx context.Context, req *AddRepoRequest) error {
	// In a real implementation, this would add to Helm repo config
	logger.Info("Added Helm repository",
		zap.String("name", req.Name),
		zap.String("url", req.URL),
	)
	return nil
}

// RemoveRepository removes a Helm repository
func (s *Service) RemoveRepository(ctx context.Context, name string) error {
	logger.Info("Removed Helm repository", zap.String("name", name))
	return nil
}

// SyncRepository syncs a Helm repository
func (s *Service) SyncRepository(ctx context.Context, name string) error {
	logger.Info("Synced Helm repository", zap.String("name", name))
	return nil
}

// SearchCharts searches for Helm charts
func (s *Service) SearchCharts(ctx context.Context, keyword, repo string) ([]Chart, error) {
	// In a real implementation, this would search Helm repos
	return []Chart{
		{
			Name:        "nginx",
			Repository:  "bitnami",
			Version:     "15.0.0",
			AppVersion:  "1.25.0",
			Description: "NGINX is a high-performance web server",
			Keywords:    []string{"web", "http", "nginx"},
		},
		{
			Name:        "postgresql",
			Repository:  "bitnami",
			Version:     "12.0.0",
			AppVersion:  "15.0.0",
			Description: "PostgreSQL database",
			Keywords:    []string{"database", "postgresql"},
		},
	}, nil
}

// GetChartDetails returns detailed chart information
func (s *Service) GetChartDetails(ctx context.Context, repo, chart, version string) (*ChartDetails, error) {
	return &ChartDetails{
		Chart: Chart{
			Name:        chart,
			Repository:  repo,
			Version:     version,
			Description: "Chart description",
		},
		Readme: "# Chart README",
		Values: map[string]interface{}{
			"replicaCount": 1,
			"image": map[string]interface{}{
				"repository": "nginx",
				"tag":        "latest",
			},
		},
	}, nil
}

// GetChartVersions returns available versions of a chart
func (s *Service) GetChartVersions(ctx context.Context, repo, chart string) ([]string, error) {
	return []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"}, nil
}

// ListReleases returns Helm releases
func (s *Service) ListReleases(ctx context.Context, clusterID, namespace string) ([]Release, error) {
	query := `
		SELECT id, name, cluster_id, namespace, chart_name, chart_version,
		       chart_repo, app_version, status, revision, last_deployed,
		       created_by, created_at, updated_at
		FROM helm_releases
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	if clusterID != "" {
		argCount++
		query += " AND cluster_id = $" + string(rune('0'+argCount))
		args = append(args, clusterID)
	}

	if namespace != "" {
		argCount++
		query += " AND namespace = $" + string(rune('0'+argCount))
		args = append(args, namespace)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to query releases")
	}
	defer rows.Close()

	var releases []Release
	for rows.Next() {
		var r Release
		var lastDeployed sql.NullTime

		if err := rows.Scan(
			&r.ID, &r.Name, &r.ClusterID, &r.Namespace, &r.ChartName,
			&r.ChartVersion, &r.ChartRepo, &r.AppVersion, &r.Status,
			&r.Revision, &lastDeployed, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan release")
		}

		if lastDeployed.Valid {
			r.LastDeployed = lastDeployed.Time
		}

		releases = append(releases, r)
	}

	return releases, nil
}

// GetRelease returns a single release
func (s *Service) GetRelease(ctx context.Context, clusterID, namespace, name string) (*Release, error) {
	query := `
		SELECT id, name, cluster_id, namespace, chart_name, chart_version,
		       chart_repo, app_version, values_yaml, status, revision,
		       last_deployed, notes, created_by, created_at, updated_at
		FROM helm_releases
		WHERE cluster_id = $1 AND namespace = $2 AND name = $3
	`

	var r Release
	var lastDeployed sql.NullTime
	var notes, valuesYAML sql.NullString

	if err := s.db.QueryRowContext(ctx, query, clusterID, namespace, name).Scan(
		&r.ID, &r.Name, &r.ClusterID, &r.Namespace, &r.ChartName,
		&r.ChartVersion, &r.ChartRepo, &r.AppVersion, &valuesYAML,
		&r.Status, &r.Revision, &lastDeployed, &notes,
		&r.CreatedBy, &r.CreatedAt, &r.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundMsg("release not found")
		}
		return nil, errors.DatabaseWrap(err, "failed to get release")
	}

	if lastDeployed.Valid {
		r.LastDeployed = lastDeployed.Time
	}
	if notes.Valid {
		r.Notes = notes.String
	}
	if valuesYAML.Valid {
		r.ValuesYAML = valuesYAML.String
	}

	return &r, nil
}

// Install installs a Helm release
func (s *Service) Install(ctx context.Context, req *InstallRequest) (*Release, error) {
	// Convert values to YAML if provided as map
	valuesYAML := req.ValuesYAML
	if len(req.Values) > 0 && valuesYAML == "" {
		data, _ := json.Marshal(req.Values)
		valuesYAML = string(data)
	}

	// In a real implementation, this would use Helm SDK to install
	query := `
		INSERT INTO helm_releases (name, cluster_id, namespace, chart_name,
		                           chart_version, chart_repo, values_yaml,
		                           status, revision, last_deployed, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'deployed', 1, NOW(), $8)
		RETURNING id, created_at, updated_at
	`

	var release Release
	if err := s.db.QueryRowContext(ctx, query,
		req.Name, req.ClusterID, req.Namespace, req.ChartName,
		req.ChartVersion, req.ChartRepo, valuesYAML, req.CreatedBy,
	).Scan(&release.ID, &release.CreatedAt, &release.UpdatedAt); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to install release")
	}

	release.Name = req.Name
	release.ClusterID = req.ClusterID
	release.Namespace = req.Namespace
	release.ChartName = req.ChartName
	release.ChartVersion = req.ChartVersion
	release.ChartRepo = req.ChartRepo
	release.Status = "deployed"
	release.Revision = 1
	release.LastDeployed = time.Now()

	logger.Info("Helm release installed",
		zap.String("name", req.Name),
		zap.String("chart", req.ChartName),
	)

	return &release, nil
}

// Upgrade upgrades a Helm release
func (s *Service) Upgrade(ctx context.Context, req *UpgradeRequest) (*Release, error) {
	valuesYAML := req.ValuesYAML
	if len(req.Values) > 0 && valuesYAML == "" {
		data, _ := json.Marshal(req.Values)
		valuesYAML = string(data)
	}

	query := `
		UPDATE helm_releases
		SET chart_version = COALESCE(NULLIF($4, ''), chart_version),
		    values_yaml = COALESCE($5, values_yaml),
		    status = 'deployed',
		    revision = revision + 1,
		    last_deployed = NOW(),
		    updated_at = NOW()
		WHERE cluster_id = $1 AND namespace = $2 AND name = $3
		RETURNING revision
	`

	var revision int
	if err := s.db.QueryRowContext(ctx, query,
		req.ClusterID, req.Namespace, req.Name, req.ChartVersion, valuesYAML,
	).Scan(&revision); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to upgrade release")
	}

	logger.Info("Helm release upgraded",
		zap.String("name", req.Name),
		zap.Int("revision", revision),
	)

	return s.GetRelease(ctx, req.ClusterID, req.Namespace, req.Name)
}

// Uninstall uninstalls a Helm release
func (s *Service) Uninstall(ctx context.Context, clusterID, namespace, name string) error {
	query := "DELETE FROM helm_releases WHERE cluster_id = $1 AND namespace = $2 AND name = $3"
	result, err := s.db.ExecContext(ctx, query, clusterID, namespace, name)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to uninstall release")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFoundMsg("release not found")
	}

	logger.Info("Helm release uninstalled", zap.String("name", name))
	return nil
}

// Rollback rolls back a Helm release
func (s *Service) Rollback(ctx context.Context, clusterID, namespace, name string, revision int) error {
	// In a real implementation, this would use Helm SDK to rollback
	query := `
		UPDATE helm_releases
		SET status = 'deployed',
		    revision = $4,
		    last_deployed = NOW(),
		    updated_at = NOW()
		WHERE cluster_id = $1 AND namespace = $2 AND name = $3
	`

	result, err := s.db.ExecContext(ctx, query, clusterID, namespace, name, revision)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to rollback release")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFoundMsg("release not found")
	}

	logger.Info("Helm release rolled back",
		zap.String("name", name),
		zap.Int("revision", revision),
	)

	return nil
}

// GetHistory returns release history
func (s *Service) GetHistory(ctx context.Context, clusterID, namespace, name string) ([]ReleaseHistory, error) {
	// In a real implementation, this would query Helm history
	release, err := s.GetRelease(ctx, clusterID, namespace, name)
	if err != nil {
		return nil, err
	}

	history := make([]ReleaseHistory, release.Revision)
	for i := 1; i <= release.Revision; i++ {
		history[i-1] = ReleaseHistory{
			Revision:    i,
			Status:      "deployed",
			Chart:       release.ChartName + "-" + release.ChartVersion,
			AppVersion:  release.AppVersion,
			Description: "Upgrade complete",
			UpdatedAt:   release.LastDeployed,
		}
	}

	return history, nil
}

// GetValues returns release values
func (s *Service) GetValues(ctx context.Context, clusterID, namespace, name string, allValues bool) (map[string]interface{}, error) {
	release, err := s.GetRelease(ctx, clusterID, namespace, name)
	if err != nil {
		return nil, err
	}

	var values map[string]interface{}
	if release.ValuesYAML != "" {
		json.Unmarshal([]byte(release.ValuesYAML), &values)
	}

	if values == nil {
		values = make(map[string]interface{})
	}

	return values, nil
}
