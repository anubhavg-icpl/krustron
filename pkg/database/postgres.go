// Package database provides database connections and utilities
// Author: Anubhav Gain <anubhavg@infopercept.com>
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/anubhavg-icpl/krustron/pkg/config"
	"github.com/anubhavg-icpl/krustron/pkg/logger"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// PostgresDB wraps the sql.DB connection
type PostgresDB struct {
	*sql.DB
	config *config.DatabaseConfig
}

// NewPostgresDB creates a new PostgreSQL connection
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL database",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
	)

	return &PostgresDB{
		DB:     db,
		config: cfg,
	}, nil
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	logger.Info("Closing PostgreSQL connection")
	return db.DB.Close()
}

// Health checks the database health
func (db *PostgresDB) Health(ctx context.Context) error {
	return db.PingContext(ctx)
}

// Transaction executes a function within a database transaction
func (db *PostgresDB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Migrate runs database migrations
func (db *PostgresDB) Migrate(ctx context.Context) error {
	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255),
			name VARCHAR(255) NOT NULL,
			avatar_url VARCHAR(512),
			provider VARCHAR(50) DEFAULT 'local',
			provider_id VARCHAR(255),
			role VARCHAR(50) DEFAULT 'user',
			is_active BOOLEAN DEFAULT true,
			last_login_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Clusters table
		`CREATE TABLE IF NOT EXISTS clusters (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) UNIQUE NOT NULL,
			display_name VARCHAR(255),
			description TEXT,
			api_server VARCHAR(512) NOT NULL,
			kubeconfig TEXT,
			auth_type VARCHAR(50) DEFAULT 'kubeconfig',
			status VARCHAR(50) DEFAULT 'pending',
			version VARCHAR(50),
			nodes_count INTEGER DEFAULT 0,
			cpu_capacity VARCHAR(50),
			memory_capacity VARCHAR(50),
			provider VARCHAR(50),
			region VARCHAR(100),
			environment VARCHAR(50) DEFAULT 'development',
			labels JSONB DEFAULT '{}',
			annotations JSONB DEFAULT '{}',
			agent_installed BOOLEAN DEFAULT false,
			agent_version VARCHAR(50),
			last_health_check TIMESTAMP WITH TIME ZONE,
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Applications table
		`CREATE TABLE IF NOT EXISTS applications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			display_name VARCHAR(255),
			description TEXT,
			cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
			namespace VARCHAR(255) NOT NULL,
			source_type VARCHAR(50) NOT NULL,
			repo_url VARCHAR(512),
			repo_branch VARCHAR(255) DEFAULT 'main',
			repo_path VARCHAR(512) DEFAULT '.',
			helm_chart VARCHAR(255),
			helm_repo VARCHAR(512),
			helm_version VARCHAR(50),
			values_yaml TEXT,
			sync_policy VARCHAR(50) DEFAULT 'manual',
			auto_sync BOOLEAN DEFAULT false,
			prune BOOLEAN DEFAULT false,
			self_heal BOOLEAN DEFAULT false,
			status VARCHAR(50) DEFAULT 'unknown',
			health_status VARCHAR(50) DEFAULT 'unknown',
			sync_status VARCHAR(50) DEFAULT 'unknown',
			last_sync_at TIMESTAMP WITH TIME ZONE,
			labels JSONB DEFAULT '{}',
			annotations JSONB DEFAULT '{}',
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(cluster_id, namespace, name)
		)`,

		// Pipelines table
		`CREATE TABLE IF NOT EXISTS pipelines (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			display_name VARCHAR(255),
			description TEXT,
			application_id UUID REFERENCES applications(id) ON DELETE CASCADE,
			trigger_type VARCHAR(50) DEFAULT 'manual',
			webhook_secret VARCHAR(255),
			cron_schedule VARCHAR(100),
			stages JSONB NOT NULL DEFAULT '[]',
			variables JSONB DEFAULT '{}',
			timeout INTEGER DEFAULT 3600,
			retry_count INTEGER DEFAULT 0,
			is_active BOOLEAN DEFAULT true,
			last_run_at TIMESTAMP WITH TIME ZONE,
			last_run_status VARCHAR(50),
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Pipeline runs table
		`CREATE TABLE IF NOT EXISTS pipeline_runs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			pipeline_id UUID REFERENCES pipelines(id) ON DELETE CASCADE,
			run_number INTEGER NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			trigger VARCHAR(50) NOT NULL,
			trigger_info JSONB DEFAULT '{}',
			stages_status JSONB DEFAULT '{}',
			current_stage VARCHAR(255),
			variables JSONB DEFAULT '{}',
			artifacts JSONB DEFAULT '[]',
			logs_url VARCHAR(512),
			started_at TIMESTAMP WITH TIME ZONE,
			finished_at TIMESTAMP WITH TIME ZONE,
			duration INTEGER,
			error_message TEXT,
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Helm releases table
		`CREATE TABLE IF NOT EXISTS helm_releases (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
			namespace VARCHAR(255) NOT NULL,
			chart_name VARCHAR(255) NOT NULL,
			chart_version VARCHAR(50) NOT NULL,
			chart_repo VARCHAR(512),
			app_version VARCHAR(50),
			values_yaml TEXT,
			status VARCHAR(50) DEFAULT 'unknown',
			revision INTEGER DEFAULT 1,
			last_deployed TIMESTAMP WITH TIME ZONE,
			notes TEXT,
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(cluster_id, namespace, name)
		)`,

		// Security scans table
		`CREATE TABLE IF NOT EXISTS security_scans (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			scan_type VARCHAR(50) NOT NULL,
			target_type VARCHAR(50) NOT NULL,
			target_id UUID NOT NULL,
			target_name VARCHAR(255) NOT NULL,
			cluster_id UUID REFERENCES clusters(id) ON DELETE SET NULL,
			status VARCHAR(50) DEFAULT 'pending',
			critical_count INTEGER DEFAULT 0,
			high_count INTEGER DEFAULT 0,
			medium_count INTEGER DEFAULT 0,
			low_count INTEGER DEFAULT 0,
			unknown_count INTEGER DEFAULT 0,
			results JSONB DEFAULT '{}',
			scanner VARCHAR(50) NOT NULL,
			scanner_version VARCHAR(50),
			started_at TIMESTAMP WITH TIME ZONE,
			finished_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Audit logs table
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE SET NULL,
			user_email VARCHAR(255),
			action VARCHAR(100) NOT NULL,
			resource_type VARCHAR(100) NOT NULL,
			resource_id VARCHAR(255),
			resource_name VARCHAR(255),
			cluster_id UUID REFERENCES clusters(id) ON DELETE SET NULL,
			cluster_name VARCHAR(255),
			old_value JSONB,
			new_value JSONB,
			metadata JSONB DEFAULT '{}',
			ip_address VARCHAR(45),
			user_agent TEXT,
			status VARCHAR(50) DEFAULT 'success',
			error_message TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// RBAC roles table
		`CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) UNIQUE NOT NULL,
			display_name VARCHAR(255),
			description TEXT,
			permissions JSONB NOT NULL DEFAULT '[]',
			is_system BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// User role assignments
		`CREATE TABLE IF NOT EXISTS user_roles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
			cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
			namespace VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id, role_id, cluster_id, namespace)
		)`,

		// Notifications table
		`CREATE TABLE IF NOT EXISTS notifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			message TEXT NOT NULL,
			type VARCHAR(50) NOT NULL,
			severity VARCHAR(50) DEFAULT 'info',
			resource_type VARCHAR(100),
			resource_id VARCHAR(255),
			is_read BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_clusters_status ON clusters(status)`,
		`CREATE INDEX IF NOT EXISTS idx_clusters_environment ON clusters(environment)`,
		`CREATE INDEX IF NOT EXISTS idx_applications_cluster ON applications(cluster_id)`,
		`CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status)`,
		`CREATE INDEX IF NOT EXISTS idx_pipelines_application ON pipelines(application_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pipeline_runs_pipeline ON pipeline_runs(pipeline_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pipeline_runs_status ON pipeline_runs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_helm_releases_cluster ON helm_releases(cluster_id)`,
		`CREATE INDEX IF NOT EXISTS idx_security_scans_target ON security_scans(target_type, target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, is_read)`,
	}

	for _, migration := range migrations {
		if _, err := db.ExecContext(ctx, migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

// SeedDefaultData seeds initial data
func (db *PostgresDB) SeedDefaultData(ctx context.Context) error {
	// Insert default roles
	roles := []struct {
		name        string
		displayName string
		description string
		permissions string
		isSystem    bool
	}{
		{
			name:        "admin",
			displayName: "Administrator",
			description: "Full access to all resources",
			permissions: `["*"]`,
			isSystem:    true,
		},
		{
			name:        "developer",
			displayName: "Developer",
			description: "Can view and deploy applications",
			permissions: `["clusters:read", "applications:*", "pipelines:*", "helm:*"]`,
			isSystem:    true,
		},
		{
			name:        "viewer",
			displayName: "Viewer",
			description: "Read-only access to all resources",
			permissions: `["*:read"]`,
			isSystem:    true,
		},
		{
			name:        "security-auditor",
			displayName: "Security Auditor",
			description: "Can view security scans and audit logs",
			permissions: `["security:read", "audit:read", "clusters:read", "applications:read"]`,
			isSystem:    true,
		},
	}

	for _, role := range roles {
		query := `
			INSERT INTO roles (name, display_name, description, permissions, is_system)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (name) DO UPDATE SET
				display_name = EXCLUDED.display_name,
				description = EXCLUDED.description,
				permissions = EXCLUDED.permissions,
				updated_at = NOW()
		`
		if _, err := db.ExecContext(ctx, query, role.name, role.displayName, role.description, role.permissions, role.isSystem); err != nil {
			return fmt.Errorf("failed to seed role %s: %w", role.name, err)
		}
	}

	logger.Info("Default data seeded successfully")
	return nil
}
