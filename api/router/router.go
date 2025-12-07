// Package router provides HTTP routing for the Krustron API
// Author: Anubhav Gain <anubhavg@infopercept.com>
package router

import (
	"net/http"
	"time"

	"github.com/anubhavg-icpl/krustron/api/handlers"
	"github.com/anubhavg-icpl/krustron/api/middleware"
	"github.com/anubhavg-icpl/krustron/internal/auth"
	"github.com/anubhavg-icpl/krustron/internal/cluster"
	"github.com/anubhavg-icpl/krustron/internal/gitops"
	"github.com/anubhavg-icpl/krustron/internal/helm"
	"github.com/anubhavg-icpl/krustron/internal/observability"
	"github.com/anubhavg-icpl/krustron/internal/pipeline"
	"github.com/anubhavg-icpl/krustron/internal/security"
	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/anubhavg-icpl/krustron/pkg/logger"
)

// Config holds router configuration
type Config struct {
	Mode        string
	CorsOrigins []string
}

// Services holds all service dependencies
type Services struct {
	Cluster       *cluster.Service
	Helm          *helm.Service
	GitOps        *gitops.Service
	Pipeline      *pipeline.Service
	Auth          *auth.Service
	Security      *security.Service
	Observability *observability.Service
}

// New creates a new Gin router
func New(cfg *Config) *gin.Engine {
	if cfg.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middleware
	r.Use(ginzap.Ginzap(logger.Get(), time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger.Get(), true))
	r.Use(middleware.RequestID())
	r.Use(middleware.RateLimiter(100, 200)) // 100 requests per second, burst 200

	// CORS
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CorsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	return r
}

// RegisterRoutes registers all API routes
func RegisterRoutes(r *gin.Engine, services *Services) {
	// Health check endpoints
	r.GET("/health", healthCheck)
	r.GET("/ready", readinessCheck(services))
	r.GET("/live", livenessCheck)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("")
		{
			public.POST("/auth/login", handlers.Login(services.Auth))
			public.POST("/auth/register", handlers.Register(services.Auth))
			public.POST("/auth/refresh", handlers.RefreshToken(services.Auth))
			public.GET("/auth/oidc/login", handlers.OIDCLogin(services.Auth))
			public.GET("/auth/oidc/callback", handlers.OIDCCallback(services.Auth))
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(services.Auth))
		{
			// Auth routes
			authRoutes := protected.Group("/auth")
			{
				authRoutes.GET("/me", handlers.GetCurrentUser(services.Auth))
				authRoutes.PUT("/me", handlers.UpdateCurrentUser(services.Auth))
				authRoutes.POST("/logout", handlers.Logout(services.Auth))
				authRoutes.PUT("/password", handlers.ChangePassword(services.Auth))
			}

			// User management (admin only)
			userRoutes := protected.Group("/users")
			userRoutes.Use(middleware.RequireRole("admin"))
			{
				userRoutes.GET("", handlers.ListUsers(services.Auth))
				userRoutes.GET("/:id", handlers.GetUser(services.Auth))
				userRoutes.POST("", handlers.CreateUser(services.Auth))
				userRoutes.PUT("/:id", handlers.UpdateUser(services.Auth))
				userRoutes.DELETE("/:id", handlers.DeleteUser(services.Auth))
				userRoutes.PUT("/:id/roles", handlers.AssignUserRoles(services.Auth))
			}

			// Cluster routes
			clusterRoutes := protected.Group("/clusters")
			{
				clusterRoutes.GET("", handlers.ListClusters(services.Cluster))
				clusterRoutes.GET("/:id", handlers.GetCluster(services.Cluster))
				clusterRoutes.POST("", handlers.CreateCluster(services.Cluster))
				clusterRoutes.PUT("/:id", handlers.UpdateCluster(services.Cluster))
				clusterRoutes.DELETE("/:id", handlers.DeleteCluster(services.Cluster))
				clusterRoutes.GET("/:id/health", handlers.GetClusterHealth(services.Cluster))
				clusterRoutes.GET("/:id/resources", handlers.GetClusterResources(services.Cluster))
				clusterRoutes.GET("/:id/namespaces", handlers.GetNamespaces(services.Cluster))
				clusterRoutes.GET("/:id/namespaces/:namespace/pods", handlers.GetPods(services.Cluster))
				clusterRoutes.GET("/:id/namespaces/:namespace/pods/:pod/logs", handlers.GetPodLogs(services.Cluster))
				clusterRoutes.GET("/:id/namespaces/:namespace/services", handlers.GetServices(services.Cluster))
				clusterRoutes.GET("/:id/namespaces/:namespace/deployments", handlers.GetDeployments(services.Cluster))
				clusterRoutes.GET("/:id/namespaces/:namespace/events", handlers.GetEvents(services.Cluster))
				clusterRoutes.POST("/:id/agent/install", handlers.InstallAgent(services.Cluster))
			}

			// Helm routes
			helmRoutes := protected.Group("/helm")
			{
				helmRoutes.GET("/repositories", handlers.ListHelmRepos(services.Helm))
				helmRoutes.POST("/repositories", handlers.AddHelmRepo(services.Helm))
				helmRoutes.DELETE("/repositories/:name", handlers.RemoveHelmRepo(services.Helm))
				helmRoutes.POST("/repositories/:name/sync", handlers.SyncHelmRepo(services.Helm))
				helmRoutes.GET("/charts", handlers.SearchCharts(services.Helm))
				helmRoutes.GET("/charts/:repo/:chart", handlers.GetChartDetails(services.Helm))
				helmRoutes.GET("/charts/:repo/:chart/versions", handlers.GetChartVersions(services.Helm))
				helmRoutes.GET("/releases", handlers.ListReleases(services.Helm))
				helmRoutes.GET("/releases/:cluster/:namespace/:name", handlers.GetRelease(services.Helm))
				helmRoutes.POST("/releases", handlers.InstallRelease(services.Helm))
				helmRoutes.PUT("/releases/:cluster/:namespace/:name", handlers.UpgradeRelease(services.Helm))
				helmRoutes.DELETE("/releases/:cluster/:namespace/:name", handlers.UninstallRelease(services.Helm))
				helmRoutes.POST("/releases/:cluster/:namespace/:name/rollback", handlers.RollbackRelease(services.Helm))
				helmRoutes.GET("/releases/:cluster/:namespace/:name/history", handlers.GetReleaseHistory(services.Helm))
				helmRoutes.GET("/releases/:cluster/:namespace/:name/values", handlers.GetReleaseValues(services.Helm))
			}

			// Application routes
			appRoutes := protected.Group("/applications")
			{
				appRoutes.GET("", handlers.ListApplications(services.GitOps))
				appRoutes.GET("/:id", handlers.GetApplication(services.GitOps))
				appRoutes.POST("", handlers.CreateApplication(services.GitOps))
				appRoutes.PUT("/:id", handlers.UpdateApplication(services.GitOps))
				appRoutes.DELETE("/:id", handlers.DeleteApplication(services.GitOps))
				appRoutes.POST("/:id/sync", handlers.SyncApplication(services.GitOps))
				appRoutes.GET("/:id/status", handlers.GetApplicationStatus(services.GitOps))
				appRoutes.GET("/:id/resources", handlers.GetApplicationResources(services.GitOps))
				appRoutes.GET("/:id/events", handlers.GetApplicationEvents(services.GitOps))
				appRoutes.GET("/:id/manifests", handlers.GetApplicationManifests(services.GitOps))
				appRoutes.GET("/:id/diff", handlers.GetApplicationDiff(services.GitOps))
			}

			// Pipeline routes
			pipelineRoutes := protected.Group("/pipelines")
			{
				pipelineRoutes.GET("", handlers.ListPipelines(services.Pipeline))
				pipelineRoutes.GET("/:id", handlers.GetPipeline(services.Pipeline))
				pipelineRoutes.POST("", handlers.CreatePipeline(services.Pipeline))
				pipelineRoutes.PUT("/:id", handlers.UpdatePipeline(services.Pipeline))
				pipelineRoutes.DELETE("/:id", handlers.DeletePipeline(services.Pipeline))
				pipelineRoutes.POST("/:id/trigger", handlers.TriggerPipeline(services.Pipeline))
				pipelineRoutes.GET("/:id/runs", handlers.ListPipelineRuns(services.Pipeline))
				pipelineRoutes.GET("/:id/runs/:runId", handlers.GetPipelineRun(services.Pipeline))
				pipelineRoutes.POST("/:id/runs/:runId/cancel", handlers.CancelPipelineRun(services.Pipeline))
				pipelineRoutes.POST("/:id/runs/:runId/retry", handlers.RetryPipelineRun(services.Pipeline))
				pipelineRoutes.GET("/:id/runs/:runId/logs", handlers.GetPipelineRunLogs(services.Pipeline))
			}

			// Security routes
			securityRoutes := protected.Group("/security")
			{
				securityRoutes.GET("/scans", handlers.ListSecurityScans(services.Security))
				securityRoutes.GET("/scans/:id", handlers.GetSecurityScan(services.Security))
				securityRoutes.POST("/scans", handlers.TriggerSecurityScan(services.Security))
				securityRoutes.GET("/vulnerabilities", handlers.ListVulnerabilities(services.Security))
				securityRoutes.GET("/policies", handlers.ListPolicies(services.Security))
				securityRoutes.POST("/policies", handlers.CreatePolicy(services.Security))
				securityRoutes.PUT("/policies/:id", handlers.UpdatePolicy(services.Security))
				securityRoutes.DELETE("/policies/:id", handlers.DeletePolicy(services.Security))
				securityRoutes.POST("/policies/:id/validate", handlers.ValidatePolicy(services.Security))
			}

			// Observability routes
			observabilityRoutes := protected.Group("/observability")
			{
				observabilityRoutes.GET("/metrics", handlers.GetMetrics(services.Observability))
				observabilityRoutes.GET("/metrics/clusters/:id", handlers.GetClusterMetrics(services.Observability))
				observabilityRoutes.GET("/metrics/applications/:id", handlers.GetApplicationMetrics(services.Observability))
				observabilityRoutes.GET("/logs", handlers.QueryLogs(services.Observability))
				observabilityRoutes.GET("/traces", handlers.QueryTraces(services.Observability))
				observabilityRoutes.GET("/alerts", handlers.ListAlerts(services.Observability))
				observabilityRoutes.GET("/dashboards", handlers.ListDashboards(services.Observability))
				observabilityRoutes.GET("/dora", handlers.GetDORAMetrics(services.Observability))
			}

			// RBAC routes
			rbacRoutes := protected.Group("/rbac")
			rbacRoutes.Use(middleware.RequireRole("admin"))
			{
				rbacRoutes.GET("/roles", handlers.ListRoles(services.Auth))
				rbacRoutes.GET("/roles/:id", handlers.GetRole(services.Auth))
				rbacRoutes.POST("/roles", handlers.CreateRole(services.Auth))
				rbacRoutes.PUT("/roles/:id", handlers.UpdateRole(services.Auth))
				rbacRoutes.DELETE("/roles/:id", handlers.DeleteRole(services.Auth))
				rbacRoutes.GET("/permissions", handlers.ListPermissions(services.Auth))
			}

			// Audit routes
			auditRoutes := protected.Group("/audit")
			auditRoutes.Use(middleware.RequireRole("admin", "security-auditor"))
			{
				auditRoutes.GET("/logs", handlers.ListAuditLogs(services.Auth))
				auditRoutes.GET("/logs/:id", handlers.GetAuditLog(services.Auth))
				auditRoutes.GET("/logs/export", handlers.ExportAuditLogs(services.Auth))
			}

			// Settings routes
			settingsRoutes := protected.Group("/settings")
			settingsRoutes.Use(middleware.RequireRole("admin"))
			{
				settingsRoutes.GET("", handlers.GetSettings(services.Auth))
				settingsRoutes.PUT("", handlers.UpdateSettings(services.Auth))
				settingsRoutes.GET("/notifications", handlers.GetNotificationSettings(services.Auth))
				settingsRoutes.PUT("/notifications", handlers.UpdateNotificationSettings(services.Auth))
			}

			// Webhooks (for external integrations)
			webhookRoutes := protected.Group("/webhooks")
			{
				webhookRoutes.POST("/github", handlers.GitHubWebhook(services.Pipeline))
				webhookRoutes.POST("/gitlab", handlers.GitLabWebhook(services.Pipeline))
				webhookRoutes.POST("/bitbucket", handlers.BitbucketWebhook(services.Pipeline))
			}
		}
	}

	// WebSocket endpoints for real-time updates
	ws := r.Group("/ws")
	ws.Use(middleware.WSAuth(services.Auth))
	{
		ws.GET("/clusters/:id/events", handlers.ClusterEventsWS(services.Cluster))
		ws.GET("/pipelines/:id/logs", handlers.PipelineLogsWS(services.Pipeline))
		ws.GET("/pods/:cluster/:namespace/:pod/logs", handlers.PodLogsWS(services.Cluster))
	}
}

// healthCheck returns a simple health check response
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "krustron",
	})
}

// readinessCheck checks if all dependencies are ready
func readinessCheck(services *Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check all service dependencies
		ready := true
		checks := make(map[string]string)

		// Add actual health checks here
		checks["database"] = "ok"
		checks["cache"] = "ok"
		checks["kubernetes"] = "ok"

		if ready {
			c.JSON(http.StatusOK, gin.H{
				"status": "ready",
				"checks": checks,
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"checks": checks,
			})
		}
	}
}

// livenessCheck returns if the service is alive
func livenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
