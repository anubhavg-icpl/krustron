// Package grpc provides gRPC server implementation for Krustron
// Author: Anubhav Gain <anubhavg@infopercept.com>
package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Config holds gRPC server configuration
type Config struct {
	Port              int
	TLSEnabled        bool
	TLSCertFile       string
	TLSKeyFile        string
	MaxRecvMsgSize    int
	MaxSendMsgSize    int
	MaxConcurrentStreams uint32
	KeepaliveTime     time.Duration
	KeepaliveTimeout  time.Duration
	EnableReflection  bool
	EnableHealthCheck bool
}

// Server represents the gRPC server
type Server struct {
	server       *grpc.Server
	logger       *zap.Logger
	config       *Config
	healthServer *health.Server
	services     map[string]interface{}
	mu           sync.RWMutex
}

// NewServer creates a new gRPC server
func NewServer(logger *zap.Logger, config *Config) (*Server, error) {
	// Set defaults
	if config.Port == 0 {
		config.Port = 9090
	}
	if config.MaxRecvMsgSize == 0 {
		config.MaxRecvMsgSize = 16 * 1024 * 1024 // 16MB
	}
	if config.MaxSendMsgSize == 0 {
		config.MaxSendMsgSize = 16 * 1024 * 1024 // 16MB
	}
	if config.MaxConcurrentStreams == 0 {
		config.MaxConcurrentStreams = 100
	}
	if config.KeepaliveTime == 0 {
		config.KeepaliveTime = 30 * time.Second
	}
	if config.KeepaliveTimeout == 0 {
		config.KeepaliveTimeout = 10 * time.Second
	}

	// Build server options
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(config.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.MaxSendMsgSize),
		grpc.MaxConcurrentStreams(config.MaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    config.KeepaliveTime,
			Timeout: config.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// Add TLS if enabled
	if config.TLSEnabled && config.TLSCertFile != "" && config.TLSKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(config.TLSCertFile, config.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Add interceptors
	opts = append(opts,
		grpc.ChainUnaryInterceptor(
			loggingUnaryInterceptor(logger),
			recoveryUnaryInterceptor(logger),
			authUnaryInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			loggingStreamInterceptor(logger),
			recoveryStreamInterceptor(logger),
			authStreamInterceptor(),
		),
	)

	server := grpc.NewServer(opts...)

	s := &Server{
		server:   server,
		logger:   logger,
		config:   config,
		services: make(map[string]interface{}),
	}

	// Register health check
	if config.EnableHealthCheck {
		s.healthServer = health.NewServer()
		grpc_health_v1.RegisterHealthServer(server, s.healthServer)
		s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	}

	// Enable reflection for debugging
	if config.EnableReflection {
		reflection.Register(server)
	}

	return s, nil
}

// RegisterService registers a gRPC service
func (s *Server) RegisterService(name string, service interface{}, registrar func(grpc.ServiceRegistrar, interface{})) {
	s.mu.Lock()
	defer s.mu.Unlock()

	registrar(s.server, service)
	s.services[name] = service

	if s.healthServer != nil {
		s.healthServer.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_SERVING)
	}

	s.logger.Info("Registered gRPC service", zap.String("service", name))
}

// Start starts the gRPC server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.logger.Info("Starting gRPC server",
		zap.String("address", addr),
		zap.Bool("tls", s.config.TLSEnabled),
	)

	return s.server.Serve(listener)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	s.logger.Info("Stopping gRPC server")
	s.server.GracefulStop()
}

// ForceStop forcefully stops the gRPC server
func (s *Server) ForceStop() {
	s.logger.Info("Force stopping gRPC server")
	s.server.Stop()
}

// SetServiceHealth sets the health status of a service
func (s *Server) SetServiceHealth(service string, healthy bool) {
	if s.healthServer != nil {
		status := grpc_health_v1.HealthCheckResponse_SERVING
		if !healthy {
			status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		}
		s.healthServer.SetServingStatus(service, status)
	}
}

// Interceptors

func loggingUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Get peer info
		var clientIP string
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		// Get metadata
		var requestID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get("x-request-id"); len(ids) > 0 {
				requestID = ids[0]
			}
		}

		// Call handler
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// Log request
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("client_ip", clientIP),
		}

		if requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			logger.Error("gRPC request failed", fields...)
		} else {
			logger.Info("gRPC request completed", fields...)
		}

		return resp, err
	}
}

func loggingStreamInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		err := handler(srv, ss)
		duration := time.Since(start)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Bool("client_stream", info.IsClientStream),
			zap.Bool("server_stream", info.IsServerStream),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			logger.Error("gRPC stream failed", fields...)
		} else {
			logger.Info("gRPC stream completed", fields...)
		}

		return err
	}
}

func recoveryUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic recovered in gRPC handler",
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

func recoveryStreamInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic recovered in gRPC stream",
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}

func authUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for health checks
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			// For now, allow unauthenticated requests
			// In production, return error
			return handler(ctx, req)
		}

		// Validate token (simplified - use proper JWT validation)
		token := tokens[0]
		if !validateToken(token) {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Add user info to context
		ctx = context.WithValue(ctx, "user_id", extractUserID(token))

		return handler(ctx, req)
	}
}

func authStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Similar to unary interceptor
		return handler(srv, ss)
	}
}

func validateToken(token string) bool {
	// Simplified validation - implement proper JWT validation
	return len(token) > 0
}

func extractUserID(token string) string {
	// Simplified extraction - implement proper JWT parsing
	return "user"
}

// Service definitions for Krustron gRPC API

// ClusterService defines cluster management operations
type ClusterService interface {
	ListClusters(ctx context.Context, req *ListClustersRequest) (*ListClustersResponse, error)
	GetCluster(ctx context.Context, req *GetClusterRequest) (*Cluster, error)
	CreateCluster(ctx context.Context, req *CreateClusterRequest) (*Cluster, error)
	UpdateCluster(ctx context.Context, req *UpdateClusterRequest) (*Cluster, error)
	DeleteCluster(ctx context.Context, req *DeleteClusterRequest) (*Empty, error)
	GetClusterHealth(ctx context.Context, req *GetClusterRequest) (*ClusterHealth, error)
	WatchClusterEvents(req *WatchEventsRequest, stream ClusterService_WatchClusterEventsServer) error
}

// ApplicationService defines application management operations
type ApplicationService interface {
	ListApplications(ctx context.Context, req *ListApplicationsRequest) (*ListApplicationsResponse, error)
	GetApplication(ctx context.Context, req *GetApplicationRequest) (*Application, error)
	CreateApplication(ctx context.Context, req *CreateApplicationRequest) (*Application, error)
	UpdateApplication(ctx context.Context, req *UpdateApplicationRequest) (*Application, error)
	DeleteApplication(ctx context.Context, req *DeleteApplicationRequest) (*Empty, error)
	SyncApplication(ctx context.Context, req *SyncApplicationRequest) (*SyncResult, error)
	WatchApplicationEvents(req *WatchEventsRequest, stream ApplicationService_WatchApplicationEventsServer) error
}

// PipelineService defines CI/CD pipeline operations
type PipelineService interface {
	ListPipelines(ctx context.Context, req *ListPipelinesRequest) (*ListPipelinesResponse, error)
	GetPipeline(ctx context.Context, req *GetPipelineRequest) (*Pipeline, error)
	CreatePipeline(ctx context.Context, req *CreatePipelineRequest) (*Pipeline, error)
	UpdatePipeline(ctx context.Context, req *UpdatePipelineRequest) (*Pipeline, error)
	DeletePipeline(ctx context.Context, req *DeletePipelineRequest) (*Empty, error)
	TriggerPipeline(ctx context.Context, req *TriggerPipelineRequest) (*PipelineRun, error)
	CancelPipeline(ctx context.Context, req *CancelPipelineRequest) (*PipelineRun, error)
	GetPipelineLogs(req *GetPipelineLogsRequest, stream PipelineService_GetPipelineLogsServer) error
}

// Message types

type Empty struct{}

type ListClustersRequest struct {
	PageSize  int32  `json:"page_size"`
	PageToken string `json:"page_token"`
	Filter    string `json:"filter"`
}

type ListClustersResponse struct {
	Clusters      []*Cluster `json:"clusters"`
	NextPageToken string     `json:"next_page_token"`
	TotalCount    int32      `json:"total_count"`
}

type GetClusterRequest struct {
	ClusterID string `json:"cluster_id"`
}

type CreateClusterRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Kubeconfig  string            `json:"kubeconfig"`
	Labels      map[string]string `json:"labels"`
}

type UpdateClusterRequest struct {
	ClusterID   string            `json:"cluster_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

type DeleteClusterRequest struct {
	ClusterID string `json:"cluster_id"`
}

type Cluster struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	NodeCount   int32             `json:"node_count"`
	Labels      map[string]string `json:"labels"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
}

type ClusterHealth struct {
	ClusterID     string            `json:"cluster_id"`
	Status        string            `json:"status"`
	NodeStatus    map[string]string `json:"node_status"`
	Components    []*ComponentHealth `json:"components"`
	LastCheckedAt int64             `json:"last_checked_at"`
}

type ComponentHealth struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type WatchEventsRequest struct {
	ResourceID string   `json:"resource_id"`
	EventTypes []string `json:"event_types"`
}

type Event struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	ResourceID string            `json:"resource_id"`
	Message    string            `json:"message"`
	Severity   string            `json:"severity"`
	Metadata   map[string]string `json:"metadata"`
	Timestamp  int64             `json:"timestamp"`
}

type ListApplicationsRequest struct {
	ClusterID string `json:"cluster_id"`
	Namespace string `json:"namespace"`
	PageSize  int32  `json:"page_size"`
	PageToken string `json:"page_token"`
}

type ListApplicationsResponse struct {
	Applications  []*Application `json:"applications"`
	NextPageToken string         `json:"next_page_token"`
	TotalCount    int32          `json:"total_count"`
}

type GetApplicationRequest struct {
	ApplicationID string `json:"application_id"`
}

type CreateApplicationRequest struct {
	Name       string            `json:"name"`
	ClusterID  string            `json:"cluster_id"`
	Namespace  string            `json:"namespace"`
	RepoURL    string            `json:"repo_url"`
	Path       string            `json:"path"`
	TargetRef  string            `json:"target_ref"`
	Labels     map[string]string `json:"labels"`
}

type UpdateApplicationRequest struct {
	ApplicationID string            `json:"application_id"`
	Name          string            `json:"name"`
	RepoURL       string            `json:"repo_url"`
	Path          string            `json:"path"`
	TargetRef     string            `json:"target_ref"`
	Labels        map[string]string `json:"labels"`
}

type DeleteApplicationRequest struct {
	ApplicationID string `json:"application_id"`
}

type SyncApplicationRequest struct {
	ApplicationID string `json:"application_id"`
	Prune         bool   `json:"prune"`
	DryRun        bool   `json:"dry_run"`
}

type Application struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	ClusterID   string            `json:"cluster_id"`
	Namespace   string            `json:"namespace"`
	RepoURL     string            `json:"repo_url"`
	Path        string            `json:"path"`
	TargetRef   string            `json:"target_ref"`
	SyncStatus  string            `json:"sync_status"`
	HealthStatus string           `json:"health_status"`
	Labels      map[string]string `json:"labels"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
}

type SyncResult struct {
	ApplicationID string   `json:"application_id"`
	Status        string   `json:"status"`
	Message       string   `json:"message"`
	Resources     []string `json:"resources"`
}

type ListPipelinesRequest struct {
	ApplicationID string `json:"application_id"`
	PageSize      int32  `json:"page_size"`
	PageToken     string `json:"page_token"`
}

type ListPipelinesResponse struct {
	Pipelines     []*Pipeline `json:"pipelines"`
	NextPageToken string      `json:"next_page_token"`
	TotalCount    int32       `json:"total_count"`
}

type GetPipelineRequest struct {
	PipelineID string `json:"pipeline_id"`
}

type CreatePipelineRequest struct {
	Name          string           `json:"name"`
	ApplicationID string           `json:"application_id"`
	Stages        []*PipelineStage `json:"stages"`
	Triggers      []*Trigger       `json:"triggers"`
}

type UpdatePipelineRequest struct {
	PipelineID string           `json:"pipeline_id"`
	Name       string           `json:"name"`
	Stages     []*PipelineStage `json:"stages"`
	Triggers   []*Trigger       `json:"triggers"`
}

type DeletePipelineRequest struct {
	PipelineID string `json:"pipeline_id"`
}

type TriggerPipelineRequest struct {
	PipelineID string            `json:"pipeline_id"`
	Parameters map[string]string `json:"parameters"`
}

type CancelPipelineRequest struct {
	PipelineRunID string `json:"pipeline_run_id"`
}

type Pipeline struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	ApplicationID string           `json:"application_id"`
	Stages        []*PipelineStage `json:"stages"`
	Triggers      []*Trigger       `json:"triggers"`
	LastRunStatus string           `json:"last_run_status"`
	CreatedAt     int64            `json:"created_at"`
	UpdatedAt     int64            `json:"updated_at"`
}

type PipelineStage struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Config  string   `json:"config"`
	DependsOn []string `json:"depends_on"`
}

type Trigger struct {
	Type   string            `json:"type"`
	Config map[string]string `json:"config"`
}

type PipelineRun struct {
	ID         string           `json:"id"`
	PipelineID string           `json:"pipeline_id"`
	Status     string           `json:"status"`
	Stages     []*StageRun      `json:"stages"`
	StartedAt  int64            `json:"started_at"`
	FinishedAt int64            `json:"finished_at"`
	TriggeredBy string          `json:"triggered_by"`
}

type StageRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	StartedAt  int64  `json:"started_at"`
	FinishedAt int64  `json:"finished_at"`
	Message    string `json:"message"`
}

type GetPipelineLogsRequest struct {
	PipelineRunID string `json:"pipeline_run_id"`
	StageName     string `json:"stage_name"`
	Follow        bool   `json:"follow"`
	TailLines     int32  `json:"tail_lines"`
}

type LogLine struct {
	Timestamp int64  `json:"timestamp"`
	Stage     string `json:"stage"`
	Message   string `json:"message"`
	Level     string `json:"level"`
}

// Stream interfaces

type ClusterService_WatchClusterEventsServer interface {
	Send(*Event) error
	grpc.ServerStream
}

type ApplicationService_WatchApplicationEventsServer interface {
	Send(*Event) error
	grpc.ServerStream
}

type PipelineService_GetPipelineLogsServer interface {
	Send(*LogLine) error
	grpc.ServerStream
}
