// Krustron - Kubernetes Native Platform
// A Devtron alternative for unified Kubernetes management
//
// Author: Anubhav Gain <anubhavg@infopercept.com>
// License: Apache 2.0
//
// @title Krustron API
// @version 1.0
// @description Kubernetes-native platform for CI/CD, GitOps, and Cluster Management
// @termsOfService http://swagger.io/terms/
//
// @contact.name Anubhav Gain
// @contact.email anubhavg@infopercept.com
//
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host localhost:8080
// @BasePath /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anubhavg-icpl/krustron/api/router"
	"github.com/anubhavg-icpl/krustron/internal/cluster"
	"github.com/anubhavg-icpl/krustron/internal/gitops"
	"github.com/anubhavg-icpl/krustron/internal/helm"
	"github.com/anubhavg-icpl/krustron/internal/pipeline"
	"github.com/anubhavg-icpl/krustron/internal/auth"
	"github.com/anubhavg-icpl/krustron/internal/security"
	"github.com/anubhavg-icpl/krustron/internal/observability"
	"github.com/anubhavg-icpl/krustron/pkg/cache"
	"github.com/anubhavg-icpl/krustron/pkg/config"
	"github.com/anubhavg-icpl/krustron/pkg/database"
	"github.com/anubhavg-icpl/krustron/pkg/kube"
	"github.com/anubhavg-icpl/krustron/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	cfgFile string
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "krustron",
		Short: "Krustron - Kubernetes Native Platform",
		Long: `Krustron is an open-source Kubernetes-native platform that combines
unified dashboard, end-to-end GitOps CI/CD pipelines, multi-cluster management,
and integrated observability/security with fine-grained RBAC and AI-assisted operations.

Built by Anubhav Gain <anubhavg@infopercept.com>`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// Add subcommands
	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(migrateCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the Krustron API server",
		Long:  "Start the Krustron API server with all modules enabled",
		RunE:  runServer,
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Krustron %s\n", version)
			fmt.Printf("  Commit: %s\n", commit)
			fmt.Printf("  Built:  %s\n", date)
		},
	}
}

func migrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			db, err := database.NewPostgresDB(&cfg.Database)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			ctx := context.Background()
			if err := db.Migrate(ctx); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			if err := db.SeedDefaultData(ctx); err != nil {
				return fmt.Errorf("failed to seed data: %w", err)
			}

			fmt.Println("Migrations completed successfully")
			return nil
		},
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	logCfg := &logger.Config{
		Level:       cfg.Logger.Level,
		Format:      cfg.Logger.Format,
		Output:      cfg.Logger.Output,
		Development: cfg.Logger.Development,
	}
	if err := logger.Init(logCfg); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("Starting Krustron",
		zap.String("version", version),
		zap.String("commit", commit),
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(ctx); err != nil {
		logger.Warn("Migration failed", zap.Error(err))
	}

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		logger.Warn("Failed to connect to Redis, continuing without cache", zap.Error(err))
	} else {
		defer redisCache.Close()
	}

	// Initialize Kubernetes client manager
	kubeManager, err := kube.NewClientManager(&cfg.Kubernetes)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client manager: %w", err)
	}

	// Try to get local cluster client
	localClient, err := kubeManager.GetLocalClient()
	if err != nil {
		logger.Warn("Failed to connect to local Kubernetes cluster", zap.Error(err))
	} else {
		logger.Info("Connected to local Kubernetes cluster",
			zap.String("version", localClient.Version),
		)
	}

	// Initialize services
	clusterService := cluster.NewService(db, kubeManager, redisCache)
	helmService := helm.NewService(db, kubeManager, redisCache)
	gitopsService := gitops.NewService(db, kubeManager, &cfg.GitOps)
	pipelineService := pipeline.NewService(db, kubeManager, redisCache, gitopsService)
	authService := auth.NewService(db, redisCache, &cfg.Auth)
	securityService := security.NewService(db, kubeManager, &cfg.Security)
	observabilityService := observability.NewService(&cfg.Observability)

	// Create router
	r := router.New(&router.Config{
		Mode:        cfg.Server.Mode,
		CorsOrigins: cfg.Server.CorsOrigins,
	})

	// Register routes
	router.RegisterRoutes(r, &router.Services{
		Cluster:       clusterService,
		Helm:          helmService,
		GitOps:        gitopsService,
		Pipeline:      pipelineService,
		Auth:          authService,
		Security:      securityService,
		Observability: observabilityService,
	})

	// Start server
	serverAddr := cfg.Server.Addr()
	logger.Info("Starting HTTP server",
		zap.String("addr", serverAddr),
		zap.String("mode", cfg.Server.Mode),
	)

	// Start HTTP server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := r.Run(serverAddr); err != nil {
			errChan <- err
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Received shutdown signal")
	case err := <-errChan:
		logger.Error("Server error", zap.Error(err))
		return err
	}

	// Graceful shutdown
	logger.Info("Shutting down server...")
	cancel()

	logger.Info("Server stopped")
	return nil
}
