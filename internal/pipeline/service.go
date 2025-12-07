// Package pipeline provides CI/CD pipeline management
// Author: Anubhav Gain <anubhavg@infopercept.com>
package pipeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/anubhavg-icpl/krustron/internal/gitops"
	"github.com/anubhavg-icpl/krustron/pkg/cache"
	"github.com/anubhavg-icpl/krustron/pkg/database"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/anubhavg-icpl/krustron/pkg/kube"
	"github.com/anubhavg-icpl/krustron/pkg/logger"
	"go.uber.org/zap"
)

// Service provides pipeline management functionality
type Service struct {
	db            *database.PostgresDB
	kubeManager   *kube.ClientManager
	cache         *cache.RedisCache
	gitopsService *gitops.Service
}

// NewService creates a new pipeline service
func NewService(db *database.PostgresDB, kubeManager *kube.ClientManager, cache *cache.RedisCache, gitopsSvc *gitops.Service) *Service {
	return &Service{
		db:            db,
		kubeManager:   kubeManager,
		cache:         cache,
		gitopsService: gitopsSvc,
	}
}

// Pipeline represents a CI/CD pipeline
type Pipeline struct {
	ID            string                 `json:"id" db:"id"`
	Name          string                 `json:"name" db:"name"`
	DisplayName   string                 `json:"display_name" db:"display_name"`
	Description   string                 `json:"description" db:"description"`
	ApplicationID string                 `json:"application_id" db:"application_id"`
	TriggerType   string                 `json:"trigger_type" db:"trigger_type"`
	WebhookSecret string                 `json:"-" db:"webhook_secret"`
	CronSchedule  string                 `json:"cron_schedule" db:"cron_schedule"`
	Stages        []Stage                `json:"stages"`
	Variables     map[string]string      `json:"variables"`
	Timeout       int                    `json:"timeout" db:"timeout"`
	RetryCount    int                    `json:"retry_count" db:"retry_count"`
	IsActive      bool                   `json:"is_active" db:"is_active"`
	LastRunAt     *time.Time             `json:"last_run_at" db:"last_run_at"`
	LastRunStatus string                 `json:"last_run_status" db:"last_run_status"`
	CreatedBy     string                 `json:"created_by" db:"created_by"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// Stage represents a pipeline stage
type Stage struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"` // build, deploy, test, security, approve
	Image    string   `json:"image,omitempty"`
	Commands []string `json:"commands,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	When     string   `json:"when,omitempty"`
	Timeout  int      `json:"timeout,omitempty"`
	Parallel bool     `json:"parallel,omitempty"`
}

// PipelineRun represents a pipeline execution
type PipelineRun struct {
	ID           string                 `json:"id" db:"id"`
	PipelineID   string                 `json:"pipeline_id" db:"pipeline_id"`
	RunNumber    int                    `json:"run_number" db:"run_number"`
	Status       string                 `json:"status" db:"status"`
	Trigger      string                 `json:"trigger" db:"trigger"`
	TriggerInfo  map[string]interface{} `json:"trigger_info"`
	StagesStatus map[string]StageStatus `json:"stages_status"`
	CurrentStage string                 `json:"current_stage" db:"current_stage"`
	Variables    map[string]string      `json:"variables"`
	Artifacts    []Artifact             `json:"artifacts"`
	LogsURL      string                 `json:"logs_url" db:"logs_url"`
	StartedAt    *time.Time             `json:"started_at" db:"started_at"`
	FinishedAt   *time.Time             `json:"finished_at" db:"finished_at"`
	Duration     int                    `json:"duration" db:"duration"`
	ErrorMessage string                 `json:"error_message,omitempty" db:"error_message"`
	CreatedBy    string                 `json:"created_by" db:"created_by"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// StageStatus represents the status of a stage
type StageStatus struct {
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	Duration   int        `json:"duration"`
	Logs       string     `json:"logs,omitempty"`
}

// Artifact represents a build artifact
type Artifact struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

// ListFilters contains filters for listing pipelines
type ListFilters struct {
	Page          int
	Limit         int
	ApplicationID string
}

// CreateRequest contains pipeline creation data
type CreateRequest struct {
	Name          string            `json:"name" binding:"required"`
	DisplayName   string            `json:"display_name"`
	Description   string            `json:"description"`
	ApplicationID string            `json:"application_id" binding:"required"`
	TriggerType   string            `json:"trigger_type" binding:"required,oneof=manual webhook cron"`
	CronSchedule  string            `json:"cron_schedule"`
	Stages        []Stage           `json:"stages" binding:"required"`
	Variables     map[string]string `json:"variables"`
	Timeout       int               `json:"timeout"`
	RetryCount    int               `json:"retry_count"`
	CreatedBy     string            `json:"-"`
}

// UpdateRequest contains pipeline update data
type UpdateRequest struct {
	DisplayName  string            `json:"display_name"`
	Description  string            `json:"description"`
	TriggerType  string            `json:"trigger_type"`
	CronSchedule string            `json:"cron_schedule"`
	Stages       []Stage           `json:"stages"`
	Variables    map[string]string `json:"variables"`
	Timeout      int               `json:"timeout"`
	RetryCount   int               `json:"retry_count"`
	IsActive     *bool             `json:"is_active"`
}

// TriggerRequest contains pipeline trigger data
type TriggerRequest struct {
	Variables   map[string]string `json:"variables"`
	TriggeredBy string            `json:"-"`
}

// List returns all pipelines with filters
func (s *Service) List(ctx context.Context, filters *ListFilters) ([]Pipeline, int, error) {
	query := `
		SELECT id, name, display_name, description, application_id,
		       trigger_type, cron_schedule, stages, variables, timeout,
		       retry_count, is_active, last_run_at, last_run_status,
		       created_by, created_at, updated_at
		FROM pipelines
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM pipelines WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if filters.ApplicationID != "" {
		argCount++
		query += " AND application_id = $1"
		countQuery += " AND application_id = $1"
		args = append(args, filters.ApplicationID)
	}

	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to count pipelines")
	}

	offset := (filters.Page - 1) * filters.Limit
	query += " ORDER BY created_at DESC LIMIT $" + string(rune('0'+argCount+1)) + " OFFSET $" + string(rune('0'+argCount+2))
	args = append(args, filters.Limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to query pipelines")
	}
	defer rows.Close()

	var pipelines []Pipeline
	for rows.Next() {
		var p Pipeline
		var stages, variables []byte
		var lastRunAt sql.NullTime
		var lastRunStatus sql.NullString

		if err := rows.Scan(
			&p.ID, &p.Name, &p.DisplayName, &p.Description, &p.ApplicationID,
			&p.TriggerType, &p.CronSchedule, &stages, &variables, &p.Timeout,
			&p.RetryCount, &p.IsActive, &lastRunAt, &lastRunStatus,
			&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, errors.DatabaseWrap(err, "failed to scan pipeline")
		}

		if lastRunAt.Valid {
			p.LastRunAt = &lastRunAt.Time
		}
		if lastRunStatus.Valid {
			p.LastRunStatus = lastRunStatus.String
		}
		json.Unmarshal(stages, &p.Stages)
		json.Unmarshal(variables, &p.Variables)

		pipelines = append(pipelines, p)
	}

	return pipelines, total, nil
}

// Get returns a single pipeline
func (s *Service) Get(ctx context.Context, id string) (*Pipeline, error) {
	query := `
		SELECT id, name, display_name, description, application_id,
		       trigger_type, cron_schedule, stages, variables, timeout,
		       retry_count, is_active, last_run_at, last_run_status,
		       created_by, created_at, updated_at
		FROM pipelines WHERE id = $1
	`

	var p Pipeline
	var stages, variables []byte
	var lastRunAt sql.NullTime
	var lastRunStatus sql.NullString

	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.DisplayName, &p.Description, &p.ApplicationID,
		&p.TriggerType, &p.CronSchedule, &stages, &variables, &p.Timeout,
		&p.RetryCount, &p.IsActive, &lastRunAt, &lastRunStatus,
		&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("pipeline", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get pipeline")
	}

	if lastRunAt.Valid {
		p.LastRunAt = &lastRunAt.Time
	}
	if lastRunStatus.Valid {
		p.LastRunStatus = lastRunStatus.String
	}
	json.Unmarshal(stages, &p.Stages)
	json.Unmarshal(variables, &p.Variables)

	return &p, nil
}

// Create creates a new pipeline
func (s *Service) Create(ctx context.Context, req *CreateRequest) (*Pipeline, error) {
	stages, _ := json.Marshal(req.Stages)
	variables, _ := json.Marshal(req.Variables)

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Name
	}

	timeout := req.Timeout
	if timeout == 0 {
		timeout = 3600
	}

	query := `
		INSERT INTO pipelines (name, display_name, description, application_id,
		                       trigger_type, cron_schedule, stages, variables,
		                       timeout, retry_count, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, is_active, created_at, updated_at
	`

	var p Pipeline
	if err := s.db.QueryRowContext(ctx, query,
		req.Name, displayName, req.Description, req.ApplicationID,
		req.TriggerType, req.CronSchedule, stages, variables,
		timeout, req.RetryCount, req.CreatedBy,
	).Scan(&p.ID, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to create pipeline")
	}

	p.Name = req.Name
	p.DisplayName = displayName
	p.Description = req.Description
	p.ApplicationID = req.ApplicationID
	p.TriggerType = req.TriggerType
	p.CronSchedule = req.CronSchedule
	p.Stages = req.Stages
	p.Variables = req.Variables
	p.Timeout = timeout
	p.RetryCount = req.RetryCount
	p.CreatedBy = req.CreatedBy

	logger.Info("Pipeline created",
		zap.String("pipeline_id", p.ID),
		zap.String("name", p.Name),
	)

	return &p, nil
}

// Update updates a pipeline
func (s *Service) Update(ctx context.Context, id string, req *UpdateRequest) (*Pipeline, error) {
	stages, _ := json.Marshal(req.Stages)
	variables, _ := json.Marshal(req.Variables)

	query := `
		UPDATE pipelines
		SET display_name = COALESCE(NULLIF($2, ''), display_name),
		    description = COALESCE(NULLIF($3, ''), description),
		    trigger_type = COALESCE(NULLIF($4, ''), trigger_type),
		    cron_schedule = COALESCE(NULLIF($5, ''), cron_schedule),
		    stages = COALESCE($6, stages),
		    variables = COALESCE($7, variables),
		    timeout = COALESCE(NULLIF($8, 0), timeout),
		    retry_count = COALESCE(NULLIF($9, 0), retry_count),
		    is_active = COALESCE($10, is_active),
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.db.ExecContext(ctx, query, id,
		req.DisplayName, req.Description, req.TriggerType, req.CronSchedule,
		stages, variables, req.Timeout, req.RetryCount, req.IsActive,
	)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to update pipeline")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.NotFound("pipeline", id)
	}

	return s.Get(ctx, id)
}

// Delete deletes a pipeline
func (s *Service) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM pipelines WHERE id = $1"
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to delete pipeline")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFound("pipeline", id)
	}

	logger.Info("Pipeline deleted", zap.String("pipeline_id", id))
	return nil
}

// Trigger triggers a pipeline run
func (s *Service) Trigger(ctx context.Context, id string, req *TriggerRequest) (*PipelineRun, error) {
	pipeline, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if !pipeline.IsActive {
		return nil, errors.BadRequest("pipeline is not active")
	}

	// Get next run number
	var runNumber int
	s.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(run_number), 0) + 1 FROM pipeline_runs WHERE pipeline_id = $1", id).Scan(&runNumber)

	// Merge variables
	variables := make(map[string]string)
	for k, v := range pipeline.Variables {
		variables[k] = v
	}
	for k, v := range req.Variables {
		variables[k] = v
	}

	variablesJSON, _ := json.Marshal(variables)
	triggerInfo, _ := json.Marshal(map[string]interface{}{"triggered_by": req.TriggeredBy})
	stagesStatus, _ := json.Marshal(map[string]StageStatus{})

	now := time.Now()

	query := `
		INSERT INTO pipeline_runs (pipeline_id, run_number, status, trigger,
		                           trigger_info, stages_status, variables,
		                           started_at, created_by)
		VALUES ($1, $2, 'running', 'manual', $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	var run PipelineRun
	if err := s.db.QueryRowContext(ctx, query,
		id, runNumber, triggerInfo, stagesStatus, variablesJSON, now, req.TriggeredBy,
	).Scan(&run.ID, &run.CreatedAt); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to create pipeline run")
	}

	run.PipelineID = id
	run.RunNumber = runNumber
	run.Status = "running"
	run.Trigger = "manual"
	run.Variables = variables
	run.StartedAt = &now
	run.CreatedBy = req.TriggeredBy

	// Update pipeline last run
	s.db.ExecContext(ctx, "UPDATE pipelines SET last_run_at = $2, last_run_status = 'running' WHERE id = $1", id, now)

	logger.Info("Pipeline triggered",
		zap.String("pipeline_id", id),
		zap.Int("run_number", runNumber),
	)

	return &run, nil
}

// ListRuns returns pipeline runs
func (s *Service) ListRuns(ctx context.Context, pipelineID string, page, limit int, status string) ([]PipelineRun, int, error) {
	query := `
		SELECT id, pipeline_id, run_number, status, trigger, trigger_info,
		       stages_status, current_stage, variables, artifacts, logs_url,
		       started_at, finished_at, duration, error_message, created_by, created_at
		FROM pipeline_runs
		WHERE pipeline_id = $1
	`
	countQuery := "SELECT COUNT(*) FROM pipeline_runs WHERE pipeline_id = $1"
	args := []interface{}{pipelineID}
	argCount := 1

	if status != "" {
		argCount++
		query += " AND status = $2"
		countQuery += " AND status = $2"
		args = append(args, status)
	}

	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to count runs")
	}

	offset := (page - 1) * limit
	query += " ORDER BY created_at DESC LIMIT $" + string(rune('0'+argCount+1)) + " OFFSET $" + string(rune('0'+argCount+2))
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to query runs")
	}
	defer rows.Close()

	var runs []PipelineRun
	for rows.Next() {
		var run PipelineRun
		var triggerInfo, stagesStatus, variables, artifacts []byte
		var startedAt, finishedAt sql.NullTime
		var currentStage, logsURL, errorMessage sql.NullString

		if err := rows.Scan(
			&run.ID, &run.PipelineID, &run.RunNumber, &run.Status, &run.Trigger,
			&triggerInfo, &stagesStatus, &currentStage, &variables, &artifacts,
			&logsURL, &startedAt, &finishedAt, &run.Duration, &errorMessage,
			&run.CreatedBy, &run.CreatedAt,
		); err != nil {
			return nil, 0, errors.DatabaseWrap(err, "failed to scan run")
		}

		if startedAt.Valid {
			run.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			run.FinishedAt = &finishedAt.Time
		}
		if currentStage.Valid {
			run.CurrentStage = currentStage.String
		}
		if logsURL.Valid {
			run.LogsURL = logsURL.String
		}
		if errorMessage.Valid {
			run.ErrorMessage = errorMessage.String
		}
		json.Unmarshal(triggerInfo, &run.TriggerInfo)
		json.Unmarshal(stagesStatus, &run.StagesStatus)
		json.Unmarshal(variables, &run.Variables)
		json.Unmarshal(artifacts, &run.Artifacts)

		runs = append(runs, run)
	}

	return runs, total, nil
}

// GetRun returns a single pipeline run
func (s *Service) GetRun(ctx context.Context, pipelineID, runID string) (*PipelineRun, error) {
	query := `
		SELECT id, pipeline_id, run_number, status, trigger, trigger_info,
		       stages_status, current_stage, variables, artifacts, logs_url,
		       started_at, finished_at, duration, error_message, created_by, created_at
		FROM pipeline_runs
		WHERE pipeline_id = $1 AND id = $2
	`

	var run PipelineRun
	var triggerInfo, stagesStatus, variables, artifacts []byte
	var startedAt, finishedAt sql.NullTime
	var currentStage, logsURL, errorMessage sql.NullString

	if err := s.db.QueryRowContext(ctx, query, pipelineID, runID).Scan(
		&run.ID, &run.PipelineID, &run.RunNumber, &run.Status, &run.Trigger,
		&triggerInfo, &stagesStatus, &currentStage, &variables, &artifacts,
		&logsURL, &startedAt, &finishedAt, &run.Duration, &errorMessage,
		&run.CreatedBy, &run.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("pipeline run", runID)
		}
		return nil, errors.DatabaseWrap(err, "failed to get run")
	}

	if startedAt.Valid {
		run.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		run.FinishedAt = &finishedAt.Time
	}
	if currentStage.Valid {
		run.CurrentStage = currentStage.String
	}
	if logsURL.Valid {
		run.LogsURL = logsURL.String
	}
	if errorMessage.Valid {
		run.ErrorMessage = errorMessage.String
	}
	json.Unmarshal(triggerInfo, &run.TriggerInfo)
	json.Unmarshal(stagesStatus, &run.StagesStatus)
	json.Unmarshal(variables, &run.Variables)
	json.Unmarshal(artifacts, &run.Artifacts)

	return &run, nil
}

// CancelRun cancels a pipeline run
func (s *Service) CancelRun(ctx context.Context, pipelineID, runID string) error {
	query := `
		UPDATE pipeline_runs
		SET status = 'cancelled',
		    finished_at = NOW(),
		    duration = EXTRACT(EPOCH FROM (NOW() - started_at))::integer
		WHERE pipeline_id = $1 AND id = $2 AND status = 'running'
	`

	result, err := s.db.ExecContext(ctx, query, pipelineID, runID)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to cancel run")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.BadRequest("run is not running or not found")
	}

	logger.Info("Pipeline run cancelled", zap.String("run_id", runID))
	return nil
}

// RetryRun retries a failed pipeline run
func (s *Service) RetryRun(ctx context.Context, pipelineID, runID, userID string) (*PipelineRun, error) {
	run, err := s.GetRun(ctx, pipelineID, runID)
	if err != nil {
		return nil, err
	}

	if run.Status != "failed" && run.Status != "cancelled" {
		return nil, errors.BadRequest("can only retry failed or cancelled runs")
	}

	return s.Trigger(ctx, pipelineID, &TriggerRequest{
		Variables:   run.Variables,
		TriggeredBy: userID,
	})
}

// GetRunLogs returns logs for a pipeline run
func (s *Service) GetRunLogs(ctx context.Context, pipelineID, runID, stage string) (string, error) {
	// In a real implementation, this would fetch logs from a storage backend
	return "Pipeline logs would appear here...", nil
}

// HandleGitHubWebhook handles GitHub webhook events
func (s *Service) HandleGitHubWebhook(ctx context.Context, event, signature string, body []byte) error {
	logger.Info("Received GitHub webhook", zap.String("event", event))
	// Process webhook and trigger pipelines
	return nil
}

// HandleGitLabWebhook handles GitLab webhook events
func (s *Service) HandleGitLabWebhook(ctx context.Context, event, token string, body []byte) error {
	logger.Info("Received GitLab webhook", zap.String("event", event))
	return nil
}

// HandleBitbucketWebhook handles Bitbucket webhook events
func (s *Service) HandleBitbucketWebhook(ctx context.Context, event string, body []byte) error {
	logger.Info("Received Bitbucket webhook", zap.String("event", event))
	return nil
}
