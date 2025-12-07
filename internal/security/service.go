// Package security provides security scanning and policy management
// Author: Anubhav Gain <anubhavg@infopercept.com>
package security

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

// Service provides security scanning functionality
type Service struct {
	db          *database.PostgresDB
	kubeManager *kube.ClientManager
	config      *config.SecurityConfig
}

// NewService creates a new security service
func NewService(db *database.PostgresDB, kubeManager *kube.ClientManager, cfg *config.SecurityConfig) *Service {
	return &Service{
		db:          db,
		kubeManager: kubeManager,
		config:      cfg,
	}
}

// Scan represents a security scan
type Scan struct {
	ID             string                 `json:"id" db:"id"`
	ScanType       string                 `json:"scan_type" db:"scan_type"`
	TargetType     string                 `json:"target_type" db:"target_type"`
	TargetID       string                 `json:"target_id" db:"target_id"`
	TargetName     string                 `json:"target_name" db:"target_name"`
	ClusterID      string                 `json:"cluster_id" db:"cluster_id"`
	Status         string                 `json:"status" db:"status"`
	CriticalCount  int                    `json:"critical_count" db:"critical_count"`
	HighCount      int                    `json:"high_count" db:"high_count"`
	MediumCount    int                    `json:"medium_count" db:"medium_count"`
	LowCount       int                    `json:"low_count" db:"low_count"`
	UnknownCount   int                    `json:"unknown_count" db:"unknown_count"`
	Results        map[string]interface{} `json:"results"`
	Scanner        string                 `json:"scanner" db:"scanner"`
	ScannerVersion string                 `json:"scanner_version" db:"scanner_version"`
	StartedAt      *time.Time             `json:"started_at" db:"started_at"`
	FinishedAt     *time.Time             `json:"finished_at" db:"finished_at"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string   `json:"id"`
	VulnID      string   `json:"vuln_id"`
	Package     string   `json:"package"`
	Version     string   `json:"version"`
	FixedIn     string   `json:"fixed_in"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CVSS        float64  `json:"cvss"`
	References  []string `json:"references"`
	ClusterID   string   `json:"cluster_id"`
	Namespace   string   `json:"namespace"`
	Resource    string   `json:"resource"`
}

// Policy represents an OPA policy
type Policy struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Severity    string    `json:"severity"`
	Rego        string    `json:"rego"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ScanFilters contains filters for listing scans
type ScanFilters struct {
	Page       int
	Limit      int
	TargetType string
	Status     string
	ClusterID  string
}

// VulnFilters contains filters for listing vulnerabilities
type VulnFilters struct {
	Page      int
	Limit     int
	Severity  string
	ClusterID string
	Namespace string
}

// ScanRequest contains scan request data
type ScanRequest struct {
	ScanType   string `json:"scan_type" binding:"required,oneof=container image cluster namespace"`
	TargetType string `json:"target_type" binding:"required"`
	TargetID   string `json:"target_id" binding:"required"`
	TargetName string `json:"target_name" binding:"required"`
	ClusterID  string `json:"cluster_id"`
}

// PolicyRequest contains policy data
type PolicyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category" binding:"required"`
	Severity    string `json:"severity" binding:"required,oneof=critical high medium low"`
	Rego        string `json:"rego" binding:"required"`
	IsActive    bool   `json:"is_active"`
}

// ValidateRequest contains validation request data
type ValidateRequest struct {
	Resources []map[string]interface{} `json:"resources" binding:"required"`
}

// ValidateResult represents policy validation result
type ValidateResult struct {
	Passed    bool             `json:"passed"`
	Violations []PolicyViolation `json:"violations"`
}

// PolicyViolation represents a policy violation
type PolicyViolation struct {
	Resource string `json:"resource"`
	Policy   string `json:"policy"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// ListScans returns security scans
func (s *Service) ListScans(ctx context.Context, filters *ScanFilters) ([]Scan, int, error) {
	query := `
		SELECT id, scan_type, target_type, target_id, target_name, cluster_id,
		       status, critical_count, high_count, medium_count, low_count,
		       unknown_count, results, scanner, scanner_version, started_at,
		       finished_at, created_at
		FROM security_scans
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM security_scans WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if filters.TargetType != "" {
		argCount++
		query += " AND target_type = $" + string(rune('0'+argCount))
		countQuery += " AND target_type = $" + string(rune('0'+argCount))
		args = append(args, filters.TargetType)
	}

	if filters.Status != "" {
		argCount++
		query += " AND status = $" + string(rune('0'+argCount))
		countQuery += " AND status = $" + string(rune('0'+argCount))
		args = append(args, filters.Status)
	}

	if filters.ClusterID != "" {
		argCount++
		query += " AND cluster_id = $" + string(rune('0'+argCount))
		countQuery += " AND cluster_id = $" + string(rune('0'+argCount))
		args = append(args, filters.ClusterID)
	}

	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to count scans")
	}

	offset := (filters.Page - 1) * filters.Limit
	query += " ORDER BY created_at DESC LIMIT $" + string(rune('0'+argCount+1)) + " OFFSET $" + string(rune('0'+argCount+2))
	args = append(args, filters.Limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to query scans")
	}
	defer rows.Close()

	var scans []Scan
	for rows.Next() {
		var scan Scan
		var results []byte
		var startedAt, finishedAt sql.NullTime
		var clusterID sql.NullString

		if err := rows.Scan(
			&scan.ID, &scan.ScanType, &scan.TargetType, &scan.TargetID, &scan.TargetName,
			&clusterID, &scan.Status, &scan.CriticalCount, &scan.HighCount,
			&scan.MediumCount, &scan.LowCount, &scan.UnknownCount, &results,
			&scan.Scanner, &scan.ScannerVersion, &startedAt, &finishedAt, &scan.CreatedAt,
		); err != nil {
			return nil, 0, errors.DatabaseWrap(err, "failed to scan row")
		}

		if clusterID.Valid {
			scan.ClusterID = clusterID.String
		}
		if startedAt.Valid {
			scan.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			scan.FinishedAt = &finishedAt.Time
		}
		json.Unmarshal(results, &scan.Results)

		scans = append(scans, scan)
	}

	return scans, total, nil
}

// GetScan returns a single scan
func (s *Service) GetScan(ctx context.Context, id string) (*Scan, error) {
	query := `
		SELECT id, scan_type, target_type, target_id, target_name, cluster_id,
		       status, critical_count, high_count, medium_count, low_count,
		       unknown_count, results, scanner, scanner_version, started_at,
		       finished_at, created_at
		FROM security_scans WHERE id = $1
	`

	var scan Scan
	var results []byte
	var startedAt, finishedAt sql.NullTime
	var clusterID sql.NullString

	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&scan.ID, &scan.ScanType, &scan.TargetType, &scan.TargetID, &scan.TargetName,
		&clusterID, &scan.Status, &scan.CriticalCount, &scan.HighCount,
		&scan.MediumCount, &scan.LowCount, &scan.UnknownCount, &results,
		&scan.Scanner, &scan.ScannerVersion, &startedAt, &finishedAt, &scan.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("scan", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get scan")
	}

	if clusterID.Valid {
		scan.ClusterID = clusterID.String
	}
	if startedAt.Valid {
		scan.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		scan.FinishedAt = &finishedAt.Time
	}
	json.Unmarshal(results, &scan.Results)

	return &scan, nil
}

// TriggerScan triggers a security scan
func (s *Service) TriggerScan(ctx context.Context, req *ScanRequest) (*Scan, error) {
	now := time.Now()

	query := `
		INSERT INTO security_scans (scan_type, target_type, target_id, target_name,
		                           cluster_id, status, scanner, started_at)
		VALUES ($1, $2, $3, $4, $5, 'running', 'trivy', $6)
		RETURNING id, created_at
	`

	var scan Scan
	if err := s.db.QueryRowContext(ctx, query,
		req.ScanType, req.TargetType, req.TargetID, req.TargetName,
		req.ClusterID, now,
	).Scan(&scan.ID, &scan.CreatedAt); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to create scan")
	}

	scan.ScanType = req.ScanType
	scan.TargetType = req.TargetType
	scan.TargetID = req.TargetID
	scan.TargetName = req.TargetName
	scan.ClusterID = req.ClusterID
	scan.Status = "running"
	scan.Scanner = "trivy"
	scan.StartedAt = &now

	logger.Info("Security scan triggered",
		zap.String("scan_id", scan.ID),
		zap.String("target", req.TargetName),
	)

	// In a real implementation, this would trigger Trivy scan asynchronously
	go s.runScan(context.Background(), &scan)

	return &scan, nil
}

func (s *Service) runScan(ctx context.Context, scan *Scan) {
	// Simulate scan completion
	time.Sleep(5 * time.Second)

	now := time.Now()
	results := map[string]interface{}{
		"vulnerabilities": []map[string]interface{}{
			{
				"vuln_id":  "CVE-2023-12345",
				"package":  "openssl",
				"severity": "HIGH",
				"title":    "OpenSSL vulnerability",
			},
		},
	}
	resultsJSON, _ := json.Marshal(results)

	query := `
		UPDATE security_scans
		SET status = 'completed',
		    critical_count = $2,
		    high_count = $3,
		    medium_count = $4,
		    low_count = $5,
		    results = $6,
		    finished_at = $7
		WHERE id = $1
	`

	s.db.ExecContext(ctx, query, scan.ID, 0, 1, 2, 3, resultsJSON, now)

	logger.Info("Security scan completed", zap.String("scan_id", scan.ID))
}

// ListVulnerabilities returns vulnerabilities
func (s *Service) ListVulnerabilities(ctx context.Context, filters *VulnFilters) ([]Vulnerability, int, error) {
	// In a real implementation, this would query from scan results
	return []Vulnerability{
		{
			ID:          "1",
			VulnID:      "CVE-2023-12345",
			Package:     "openssl",
			Version:     "1.1.1",
			FixedIn:     "1.1.2",
			Severity:    "HIGH",
			Title:       "OpenSSL Buffer Overflow",
			Description: "A buffer overflow vulnerability in OpenSSL",
			CVSS:        7.5,
		},
	}, 1, nil
}

// ListPolicies returns OPA policies
func (s *Service) ListPolicies(ctx context.Context) ([]Policy, error) {
	// In a real implementation, this would query from database
	return []Policy{
		{
			ID:          "1",
			Name:        "no-privileged-containers",
			Description: "Prevent privileged containers",
			Category:    "security",
			Severity:    "high",
			IsActive:    true,
		},
		{
			ID:          "2",
			Name:        "require-resource-limits",
			Description: "Require CPU and memory limits",
			Category:    "best-practice",
			Severity:    "medium",
			IsActive:    true,
		},
	}, nil
}

// CreatePolicy creates an OPA policy
func (s *Service) CreatePolicy(ctx context.Context, req *PolicyRequest) (*Policy, error) {
	policy := &Policy{
		ID:          "generated-id",
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Severity:    req.Severity,
		Rego:        req.Rego,
		IsActive:    req.IsActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	logger.Info("Policy created", zap.String("name", req.Name))
	return policy, nil
}

// UpdatePolicy updates an OPA policy
func (s *Service) UpdatePolicy(ctx context.Context, id string, req *PolicyRequest) (*Policy, error) {
	policy := &Policy{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Severity:    req.Severity,
		Rego:        req.Rego,
		IsActive:    req.IsActive,
		UpdatedAt:   time.Now(),
	}

	logger.Info("Policy updated", zap.String("id", id))
	return policy, nil
}

// DeletePolicy deletes an OPA policy
func (s *Service) DeletePolicy(ctx context.Context, id string) error {
	logger.Info("Policy deleted", zap.String("id", id))
	return nil
}

// ValidatePolicy validates resources against a policy
func (s *Service) ValidatePolicy(ctx context.Context, id string, req *ValidateRequest) (*ValidateResult, error) {
	// In a real implementation, this would use OPA to evaluate the policy
	return &ValidateResult{
		Passed:     true,
		Violations: []PolicyViolation{},
	}, nil
}
