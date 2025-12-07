// Package cluster provides cluster management functionality
// Author: Anubhav Gain <anubhavg@infopercept.com>
package cluster

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anubhavg-icpl/krustron/pkg/cache"
	"github.com/anubhavg-icpl/krustron/pkg/database"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/anubhavg-icpl/krustron/pkg/kube"
	"github.com/anubhavg-icpl/krustron/pkg/logger"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Service provides cluster management functionality
type Service struct {
	db          *database.PostgresDB
	kubeManager *kube.ClientManager
	cache       *cache.RedisCache
}

// NewService creates a new cluster service
func NewService(db *database.PostgresDB, kubeManager *kube.ClientManager, cache *cache.RedisCache) *Service {
	return &Service{
		db:          db,
		kubeManager: kubeManager,
		cache:       cache,
	}
}

// Cluster represents a Kubernetes cluster
type Cluster struct {
	ID              string            `json:"id" db:"id"`
	Name            string            `json:"name" db:"name"`
	DisplayName     string            `json:"display_name" db:"display_name"`
	Description     string            `json:"description" db:"description"`
	APIServer       string            `json:"api_server" db:"api_server"`
	Kubeconfig      string            `json:"-" db:"kubeconfig"`
	AuthType        string            `json:"auth_type" db:"auth_type"`
	Status          string            `json:"status" db:"status"`
	Version         string            `json:"version" db:"version"`
	NodesCount      int               `json:"nodes_count" db:"nodes_count"`
	CPUCapacity     string            `json:"cpu_capacity" db:"cpu_capacity"`
	MemoryCapacity  string            `json:"memory_capacity" db:"memory_capacity"`
	Provider        string            `json:"provider" db:"provider"`
	Region          string            `json:"region" db:"region"`
	Environment     string            `json:"environment" db:"environment"`
	Labels          map[string]string `json:"labels" db:"labels"`
	Annotations     map[string]string `json:"annotations" db:"annotations"`
	AgentInstalled  bool              `json:"agent_installed" db:"agent_installed"`
	AgentVersion    string            `json:"agent_version" db:"agent_version"`
	LastHealthCheck *time.Time        `json:"last_health_check" db:"last_health_check"`
	CreatedBy       string            `json:"created_by" db:"created_by"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}

// ListFilters contains filters for listing clusters
type ListFilters struct {
	Page        int
	Limit       int
	Environment string
	Status      string
	Provider    string
	Search      string
}

// CreateRequest contains data for creating a cluster
type CreateRequest struct {
	Name        string            `json:"name" binding:"required"`
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	APIServer   string            `json:"api_server"`
	Kubeconfig  string            `json:"kubeconfig"`
	AuthType    string            `json:"auth_type"`
	Provider    string            `json:"provider"`
	Region      string            `json:"region"`
	Environment string            `json:"environment"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedBy   string            `json:"-"`
}

// UpdateRequest contains data for updating a cluster
type UpdateRequest struct {
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Environment string            `json:"environment"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// List returns all clusters with filters
func (s *Service) List(ctx context.Context, filters *ListFilters) ([]Cluster, int, error) {
	query := `
		SELECT id, name, display_name, description, api_server, auth_type, status,
		       version, nodes_count, cpu_capacity, memory_capacity, provider, region,
		       environment, labels, annotations, agent_installed, agent_version,
		       last_health_check, created_by, created_at, updated_at
		FROM clusters
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM clusters WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if filters.Environment != "" {
		argCount++
		query += fmt.Sprintf(" AND environment = $%d", argCount)
		countQuery += fmt.Sprintf(" AND environment = $%d", argCount)
		args = append(args, filters.Environment)
	}

	if filters.Status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
	}

	if filters.Provider != "" {
		argCount++
		query += fmt.Sprintf(" AND provider = $%d", argCount)
		countQuery += fmt.Sprintf(" AND provider = $%d", argCount)
		args = append(args, filters.Provider)
	}

	// Get total count
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to count clusters")
	}

	// Add pagination
	offset := (filters.Page - 1) * filters.Limit
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, filters.Limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to query clusters")
	}
	defer rows.Close()

	var clusters []Cluster
	for rows.Next() {
		var c Cluster
		var labels, annotations []byte
		var lastHealthCheck sql.NullTime

		if err := rows.Scan(
			&c.ID, &c.Name, &c.DisplayName, &c.Description, &c.APIServer, &c.AuthType,
			&c.Status, &c.Version, &c.NodesCount, &c.CPUCapacity, &c.MemoryCapacity,
			&c.Provider, &c.Region, &c.Environment, &labels, &annotations,
			&c.AgentInstalled, &c.AgentVersion, &lastHealthCheck, &c.CreatedBy,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, 0, errors.DatabaseWrap(err, "failed to scan cluster")
		}

		if lastHealthCheck.Valid {
			c.LastHealthCheck = &lastHealthCheck.Time
		}

		json.Unmarshal(labels, &c.Labels)
		json.Unmarshal(annotations, &c.Annotations)

		clusters = append(clusters, c)
	}

	return clusters, total, nil
}

// Get returns a single cluster by ID
func (s *Service) Get(ctx context.Context, id string) (*Cluster, error) {
	// Try cache first
	if s.cache != nil {
		var cached Cluster
		if err := s.cache.Get(ctx, cache.BuildKey(cache.PrefixCluster, id), &cached); err == nil {
			return &cached, nil
		}
	}

	query := `
		SELECT id, name, display_name, description, api_server, auth_type, status,
		       version, nodes_count, cpu_capacity, memory_capacity, provider, region,
		       environment, labels, annotations, agent_installed, agent_version,
		       last_health_check, created_by, created_at, updated_at
		FROM clusters WHERE id = $1
	`

	var c Cluster
	var labels, annotations []byte
	var lastHealthCheck sql.NullTime

	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.DisplayName, &c.Description, &c.APIServer, &c.AuthType,
		&c.Status, &c.Version, &c.NodesCount, &c.CPUCapacity, &c.MemoryCapacity,
		&c.Provider, &c.Region, &c.Environment, &labels, &annotations,
		&c.AgentInstalled, &c.AgentVersion, &lastHealthCheck, &c.CreatedBy,
		&c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("cluster", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get cluster")
	}

	if lastHealthCheck.Valid {
		c.LastHealthCheck = &lastHealthCheck.Time
	}

	json.Unmarshal(labels, &c.Labels)
	json.Unmarshal(annotations, &c.Annotations)

	// Cache the result
	if s.cache != nil {
		s.cache.Set(ctx, cache.BuildKey(cache.PrefixCluster, id), &c, 5*time.Minute)
	}

	return &c, nil
}

// Create creates a new cluster
func (s *Service) Create(ctx context.Context, req *CreateRequest) (*Cluster, error) {
	// Validate kubeconfig and get cluster info
	var version string
	var nodesCount int
	if req.Kubeconfig != "" {
		client, err := s.kubeManager.AddCluster(req.Name, []byte(req.Kubeconfig))
		if err != nil {
			return nil, errors.ClusterWrap(err, "failed to connect to cluster")
		}
		version = client.Version

		// Get nodes count
		info, err := client.GetClusterInfo(ctx)
		if err == nil {
			nodesCount = info.NodesCount
		}
	}

	labels, _ := json.Marshal(req.Labels)
	annotations, _ := json.Marshal(req.Annotations)

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Name
	}

	query := `
		INSERT INTO clusters (name, display_name, description, api_server, kubeconfig,
		                      auth_type, status, version, nodes_count, provider, region,
		                      environment, labels, annotations, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`

	status := "connected"
	if version == "" {
		status = "pending"
	}

	var cluster Cluster
	if err := s.db.QueryRowContext(ctx, query,
		req.Name, displayName, req.Description, req.APIServer, req.Kubeconfig,
		req.AuthType, status, version, nodesCount, req.Provider, req.Region,
		req.Environment, labels, annotations, req.CreatedBy,
	).Scan(&cluster.ID, &cluster.CreatedAt, &cluster.UpdatedAt); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to create cluster")
	}

	cluster.Name = req.Name
	cluster.DisplayName = displayName
	cluster.Description = req.Description
	cluster.APIServer = req.APIServer
	cluster.AuthType = req.AuthType
	cluster.Status = status
	cluster.Version = version
	cluster.NodesCount = nodesCount
	cluster.Provider = req.Provider
	cluster.Region = req.Region
	cluster.Environment = req.Environment
	cluster.Labels = req.Labels
	cluster.Annotations = req.Annotations
	cluster.CreatedBy = req.CreatedBy

	logger.Info("Cluster created",
		zap.String("cluster_id", cluster.ID),
		zap.String("name", cluster.Name),
	)

	return &cluster, nil
}

// Update updates a cluster
func (s *Service) Update(ctx context.Context, id string, req *UpdateRequest) (*Cluster, error) {
	labels, _ := json.Marshal(req.Labels)
	annotations, _ := json.Marshal(req.Annotations)

	query := `
		UPDATE clusters
		SET display_name = COALESCE(NULLIF($2, ''), display_name),
		    description = COALESCE(NULLIF($3, ''), description),
		    environment = COALESCE(NULLIF($4, ''), environment),
		    labels = COALESCE($5, labels),
		    annotations = COALESCE($6, annotations),
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.db.ExecContext(ctx, query, id, req.DisplayName, req.Description,
		req.Environment, labels, annotations)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to update cluster")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.NotFound("cluster", id)
	}

	// Invalidate cache
	if s.cache != nil {
		s.cache.Delete(ctx, cache.BuildKey(cache.PrefixCluster, id))
	}

	return s.Get(ctx, id)
}

// Delete deletes a cluster
func (s *Service) Delete(ctx context.Context, id string) error {
	// Get cluster name to remove from kube manager
	cluster, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Remove from kube manager
	s.kubeManager.RemoveCluster(cluster.Name)

	query := "DELETE FROM clusters WHERE id = $1"
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to delete cluster")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFound("cluster", id)
	}

	// Invalidate cache
	if s.cache != nil {
		s.cache.Delete(ctx, cache.BuildKey(cache.PrefixCluster, id))
	}

	logger.Info("Cluster deleted", zap.String("cluster_id", id))
	return nil
}

// GetHealth returns cluster health information
func (s *Service) GetHealth(ctx context.Context, id string) (*HealthStatus, error) {
	cluster, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	client, err := s.kubeManager.GetClient(cluster.Name)
	if err != nil {
		return &HealthStatus{
			Status:  "disconnected",
			Message: err.Error(),
		}, nil
	}

	if err := client.CheckHealth(ctx); err != nil {
		return &HealthStatus{
			Status:  "unhealthy",
			Message: err.Error(),
		}, nil
	}

	info, err := client.GetClusterInfo(ctx)
	if err != nil {
		return &HealthStatus{
			Status:  "unknown",
			Message: err.Error(),
		}, nil
	}

	// Update cluster info in database
	s.updateClusterInfo(ctx, id, info)

	return &HealthStatus{
		Status:         "healthy",
		Version:        info.Version,
		NodesCount:     info.NodesCount,
		PodsCount:      info.PodsCount,
		RunningPods:    info.RunningPods,
		PendingPods:    info.PendingPods,
		FailedPods:     info.FailedPods,
		CPUCapacity:    info.CPUCapacity,
		MemoryCapacity: info.MemoryCapacity,
	}, nil
}

// HealthStatus represents cluster health
type HealthStatus struct {
	Status         string `json:"status"`
	Message        string `json:"message,omitempty"`
	Version        string `json:"version,omitempty"`
	NodesCount     int    `json:"nodes_count,omitempty"`
	PodsCount      int    `json:"pods_count,omitempty"`
	RunningPods    int    `json:"running_pods,omitempty"`
	PendingPods    int    `json:"pending_pods,omitempty"`
	FailedPods     int    `json:"failed_pods,omitempty"`
	CPUCapacity    string `json:"cpu_capacity,omitempty"`
	MemoryCapacity string `json:"memory_capacity,omitempty"`
}

func (s *Service) updateClusterInfo(ctx context.Context, id string, info *kube.ClusterInfo) {
	query := `
		UPDATE clusters
		SET version = $2, nodes_count = $3, cpu_capacity = $4, memory_capacity = $5,
		    status = 'connected', last_health_check = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	s.db.ExecContext(ctx, query, id, info.Version, info.NodesCount, info.CPUCapacity, info.MemoryCapacity)

	// Invalidate cache
	if s.cache != nil {
		s.cache.Delete(ctx, cache.BuildKey(cache.PrefixCluster, id))
	}
}

// GetResources returns cluster resources summary
func (s *Service) GetResources(ctx context.Context, id string) (*ResourcesSummary, error) {
	cluster, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	client, err := s.kubeManager.GetClient(cluster.Name)
	if err != nil {
		return nil, errors.ClusterWrap(err, "failed to get cluster client")
	}

	info, err := client.GetClusterInfo(ctx)
	if err != nil {
		return nil, errors.ClusterWrap(err, "failed to get cluster info")
	}

	return &ResourcesSummary{
		Nodes:      info.NodesCount,
		Namespaces: info.NamespacesCount,
		Pods:       info.PodsCount,
		Running:    info.RunningPods,
		Pending:    info.PendingPods,
		Failed:     info.FailedPods,
	}, nil
}

// ResourcesSummary represents cluster resources
type ResourcesSummary struct {
	Nodes      int `json:"nodes"`
	Namespaces int `json:"namespaces"`
	Pods       int `json:"pods"`
	Running    int `json:"running"`
	Pending    int `json:"pending"`
	Failed     int `json:"failed"`
}

// GetNamespaces returns namespaces in a cluster
func (s *Service) GetNamespaces(ctx context.Context, id string) ([]NamespaceInfo, error) {
	cluster, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	client, err := s.kubeManager.GetClient(cluster.Name)
	if err != nil {
		return nil, errors.ClusterWrap(err, "failed to get cluster client")
	}

	namespaces, err := client.GetNamespaces(ctx)
	if err != nil {
		return nil, errors.KubernetesWrap(err, "failed to list namespaces")
	}

	result := make([]NamespaceInfo, len(namespaces))
	for i, ns := range namespaces {
		result[i] = NamespaceInfo{
			Name:      ns.Name,
			Status:    string(ns.Status.Phase),
			Labels:    ns.Labels,
			CreatedAt: ns.CreationTimestamp.Time,
		}
	}

	return result, nil
}

// NamespaceInfo represents namespace information
type NamespaceInfo struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
}

// GetPods returns pods in a namespace
func (s *Service) GetPods(ctx context.Context, clusterID, namespace string) ([]PodInfo, error) {
	cluster, err := s.Get(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	client, err := s.kubeManager.GetClient(cluster.Name)
	if err != nil {
		return nil, errors.ClusterWrap(err, "failed to get cluster client")
	}

	pods, err := client.GetPods(ctx, namespace)
	if err != nil {
		return nil, errors.KubernetesWrap(err, "failed to list pods")
	}

	result := make([]PodInfo, len(pods))
	for i, pod := range pods {
		containers := make([]ContainerInfo, len(pod.Spec.Containers))
		for j, c := range pod.Spec.Containers {
			containers[j] = ContainerInfo{
				Name:  c.Name,
				Image: c.Image,
			}
		}

		result[i] = PodInfo{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Status:     string(pod.Status.Phase),
			Node:       pod.Spec.NodeName,
			IP:         pod.Status.PodIP,
			Containers: containers,
			Restarts:   getPodRestarts(pod),
			CreatedAt:  pod.CreationTimestamp.Time,
		}
	}

	return result, nil
}

// PodInfo represents pod information
type PodInfo struct {
	Name       string          `json:"name"`
	Namespace  string          `json:"namespace"`
	Status     string          `json:"status"`
	Node       string          `json:"node"`
	IP         string          `json:"ip"`
	Containers []ContainerInfo `json:"containers"`
	Restarts   int32           `json:"restarts"`
	CreatedAt  time.Time       `json:"created_at"`
}

// ContainerInfo represents container information
type ContainerInfo struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

func getPodRestarts(pod corev1.Pod) int32 {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}

// GetPodLogs returns logs from a pod
func (s *Service) GetPodLogs(ctx context.Context, clusterID, namespace, podName, container string, tailLines int64) (string, error) {
	cluster, err := s.Get(ctx, clusterID)
	if err != nil {
		return "", err
	}

	client, err := s.kubeManager.GetClient(cluster.Name)
	if err != nil {
		return "", errors.ClusterWrap(err, "failed to get cluster client")
	}

	logs, err := client.GetPodLogs(ctx, namespace, podName, container, tailLines)
	if err != nil {
		return "", errors.KubernetesWrap(err, "failed to get pod logs")
	}

	return logs, nil
}

// GetServices returns services in a namespace
func (s *Service) GetServices(ctx context.Context, clusterID, namespace string) ([]ServiceInfo, error) {
	cluster, err := s.Get(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	client, err := s.kubeManager.GetClient(cluster.Name)
	if err != nil {
		return nil, errors.ClusterWrap(err, "failed to get cluster client")
	}

	services, err := client.GetServices(ctx, namespace)
	if err != nil {
		return nil, errors.KubernetesWrap(err, "failed to list services")
	}

	result := make([]ServiceInfo, len(services))
	for i, svc := range services {
		ports := make([]PortInfo, len(svc.Spec.Ports))
		for j, p := range svc.Spec.Ports {
			ports[j] = PortInfo{
				Name:       p.Name,
				Port:       p.Port,
				TargetPort: p.TargetPort.IntValue(),
				Protocol:   string(p.Protocol),
			}
		}

		result[i] = ServiceInfo{
			Name:       svc.Name,
			Namespace:  svc.Namespace,
			Type:       string(svc.Spec.Type),
			ClusterIP:  svc.Spec.ClusterIP,
			ExternalIP: getExternalIP(svc),
			Ports:      ports,
			CreatedAt:  svc.CreationTimestamp.Time,
		}
	}

	return result, nil
}

// ServiceInfo represents service information
type ServiceInfo struct {
	Name       string     `json:"name"`
	Namespace  string     `json:"namespace"`
	Type       string     `json:"type"`
	ClusterIP  string     `json:"cluster_ip"`
	ExternalIP string     `json:"external_ip,omitempty"`
	Ports      []PortInfo `json:"ports"`
	CreatedAt  time.Time  `json:"created_at"`
}

// PortInfo represents port information
type PortInfo struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort int    `json:"target_port"`
	Protocol   string `json:"protocol"`
}

func getExternalIP(svc corev1.Service) string {
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		ing := svc.Status.LoadBalancer.Ingress[0]
		if ing.IP != "" {
			return ing.IP
		}
		return ing.Hostname
	}
	if len(svc.Spec.ExternalIPs) > 0 {
		return svc.Spec.ExternalIPs[0]
	}
	return ""
}

// GetDeployments returns deployments in a namespace
func (s *Service) GetDeployments(ctx context.Context, clusterID, namespace string) ([]DeploymentInfo, error) {
	cluster, err := s.Get(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	client, err := s.kubeManager.GetClient(cluster.Name)
	if err != nil {
		return nil, errors.ClusterWrap(err, "failed to get cluster client")
	}

	deployments, err := client.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.KubernetesWrap(err, "failed to list deployments")
	}

	result := make([]DeploymentInfo, len(deployments.Items))
	for i, dep := range deployments.Items {
		result[i] = DeploymentInfo{
			Name:            dep.Name,
			Namespace:       dep.Namespace,
			Replicas:        *dep.Spec.Replicas,
			ReadyReplicas:   dep.Status.ReadyReplicas,
			UpdatedReplicas: dep.Status.UpdatedReplicas,
			Strategy:        string(dep.Spec.Strategy.Type),
			Labels:          dep.Labels,
			CreatedAt:       dep.CreationTimestamp.Time,
		}
	}

	return result, nil
}

// DeploymentInfo represents deployment information
type DeploymentInfo struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Replicas        int32             `json:"replicas"`
	ReadyReplicas   int32             `json:"ready_replicas"`
	UpdatedReplicas int32             `json:"updated_replicas"`
	Strategy        string            `json:"strategy"`
	Labels          map[string]string `json:"labels"`
	CreatedAt       time.Time         `json:"created_at"`
}

// GetEvents returns events in a namespace
func (s *Service) GetEvents(ctx context.Context, clusterID, namespace string) ([]EventInfo, error) {
	cluster, err := s.Get(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	client, err := s.kubeManager.GetClient(cluster.Name)
	if err != nil {
		return nil, errors.ClusterWrap(err, "failed to get cluster client")
	}

	events, err := client.GetEvents(ctx, namespace)
	if err != nil {
		return nil, errors.KubernetesWrap(err, "failed to list events")
	}

	result := make([]EventInfo, len(events))
	for i, event := range events {
		result[i] = EventInfo{
			Name:      event.Name,
			Namespace: event.Namespace,
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Object:    event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name,
			Count:     event.Count,
			FirstSeen: event.FirstTimestamp.Time,
			LastSeen:  event.LastTimestamp.Time,
		}
	}

	return result, nil
}

// EventInfo represents event information
type EventInfo struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Object    string    `json:"object"`
	Count     int32     `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// InstallAgent installs the Krustron agent on a cluster
func (s *Service) InstallAgent(ctx context.Context, id string) error {
	cluster, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	client, err := s.kubeManager.GetClient(cluster.Name)
	if err != nil {
		return errors.ClusterWrap(err, "failed to get cluster client")
	}

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "krustron-system",
		},
	}
	_, err = client.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		logger.Warn("Namespace already exists", zap.Error(err))
	}

	// Create agent deployment
	agent := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "krustron-agent",
			Namespace: "krustron-system",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "krustron-agent"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "krustron-agent"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "agent",
							Image: "ghcr.io/anubhavg-icpl/krustron-agent:latest",
							Env: []corev1.EnvVar{
								{Name: "CLUSTER_ID", Value: cluster.ID},
								{Name: "CLUSTER_NAME", Value: cluster.Name},
							},
						},
					},
				},
			},
		},
	}

	_, err = client.Clientset.AppsV1().Deployments("krustron-system").Create(ctx, agent, metav1.CreateOptions{})
	if err != nil {
		return errors.KubernetesWrap(err, "failed to create agent deployment")
	}

	// Update cluster status
	query := "UPDATE clusters SET agent_installed = true, agent_version = 'latest', updated_at = NOW() WHERE id = $1"
	s.db.ExecContext(ctx, query, id)

	// Invalidate cache
	if s.cache != nil {
		s.cache.Delete(ctx, cache.BuildKey(cache.PrefixCluster, id))
	}

	logger.Info("Agent installed", zap.String("cluster_id", id))
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
