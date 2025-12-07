// Package kube provides Kubernetes client utilities
// Author: Anubhav Gain <anubhavg@infopercept.com>
package kube

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/anubhavg-icpl/krustron/pkg/config"
	"github.com/anubhavg-icpl/krustron/pkg/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientManager manages Kubernetes clients for multiple clusters
type ClientManager struct {
	mu       sync.RWMutex
	clients  map[string]*ClusterClient
	config   *config.KubernetesConfig
	scheme   *runtime.Scheme
}

// ClusterClient wraps Kubernetes clients for a single cluster
type ClusterClient struct {
	Name          string
	Config        *rest.Config
	Clientset     kubernetes.Interface
	DynamicClient dynamic.Interface
	Connected     bool
	Version       string
	LastHealthAt  time.Time
}

// NewClientManager creates a new Kubernetes client manager
func NewClientManager(cfg *config.KubernetesConfig) (*ClientManager, error) {
	return &ClientManager{
		clients: make(map[string]*ClusterClient),
		config:  cfg,
		scheme:  runtime.NewScheme(),
	}, nil
}

// GetLocalClient creates a client for the local cluster
func (m *ClientManager) GetLocalClient() (*ClusterClient, error) {
	var restConfig *rest.Config
	var err error

	if m.config.InCluster {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	} else {
		kubeconfig := m.config.KubeconfigPath
		if kubeconfig == "" {
			if home := os.Getenv("HOME"); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		}

		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
		}
	}

	restConfig.QPS = m.config.QPS
	restConfig.Burst = m.config.Burst

	return m.createClient("local", restConfig)
}

// AddCluster adds a cluster by kubeconfig content
func (m *ClientManager) AddCluster(name string, kubeconfigData []byte) (*ClusterClient, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	config.QPS = m.config.QPS
	config.Burst = m.config.Burst

	client, err := m.createClient(name, config)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.clients[name] = client
	m.mu.Unlock()

	logger.Info("Added cluster", zap.String("cluster", name))
	return client, nil
}

// AddClusterByAPIServer adds a cluster by API server URL and token
func (m *ClientManager) AddClusterByAPIServer(name, apiServer, token, caCert string) (*ClusterClient, error) {
	config := &rest.Config{
		Host:        apiServer,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: []byte(caCert),
		},
		QPS:   m.config.QPS,
		Burst: m.config.Burst,
	}

	client, err := m.createClient(name, config)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.clients[name] = client
	m.mu.Unlock()

	logger.Info("Added cluster by API server", zap.String("cluster", name), zap.String("api_server", apiServer))
	return client, nil
}

// RemoveCluster removes a cluster
func (m *ClientManager) RemoveCluster(name string) {
	m.mu.Lock()
	delete(m.clients, name)
	m.mu.Unlock()
	logger.Info("Removed cluster", zap.String("cluster", name))
}

// GetClient gets a cluster client by name
func (m *ClientManager) GetClient(name string) (*ClusterClient, error) {
	m.mu.RLock()
	client, exists := m.clients[name]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("cluster %s not found", name)
	}
	return client, nil
}

// ListClusters lists all registered clusters
func (m *ClientManager) ListClusters() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.clients))
	for name := range m.clients {
		names = append(names, name)
	}
	return names
}

// createClient creates a new ClusterClient
func (m *ClientManager) createClient(name string, config *rest.Config) (*ClusterClient, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	client := &ClusterClient{
		Name:          name,
		Config:        config,
		Clientset:     clientset,
		DynamicClient: dynamicClient,
	}

	// Get cluster version
	if version, err := clientset.Discovery().ServerVersion(); err == nil {
		client.Version = version.GitVersion
		client.Connected = true
		client.LastHealthAt = time.Now()
	}

	return client, nil
}

// CheckHealth checks the health of a cluster
func (c *ClusterClient) CheckHealth(ctx context.Context) error {
	_, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		c.Connected = false
		return fmt.Errorf("cluster health check failed: %w", err)
	}

	c.Connected = true
	c.LastHealthAt = time.Now()
	return nil
}

// GetClusterInfo returns cluster information
func (c *ClusterClient) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	info := &ClusterInfo{
		Name:      c.Name,
		Version:   c.Version,
		Connected: c.Connected,
	}

	// Get nodes
	nodes, err := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	info.NodesCount = len(nodes.Items)

	// Calculate capacity
	var cpuCapacity, memoryCapacity int64
	for _, node := range nodes.Items {
		if cpu, ok := node.Status.Capacity[corev1.ResourceCPU]; ok {
			cpuCapacity += cpu.MilliValue()
		}
		if memory, ok := node.Status.Capacity[corev1.ResourceMemory]; ok {
			memoryCapacity += memory.Value()
		}
	}
	info.CPUCapacity = fmt.Sprintf("%dm", cpuCapacity)
	info.MemoryCapacity = fmt.Sprintf("%dGi", memoryCapacity/(1024*1024*1024))

	// Get namespaces count
	namespaces, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err == nil {
		info.NamespacesCount = len(namespaces.Items)
	}

	// Get pods count
	pods, err := c.Clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err == nil {
		info.PodsCount = len(pods.Items)
		for _, pod := range pods.Items {
			switch pod.Status.Phase {
			case corev1.PodRunning:
				info.RunningPods++
			case corev1.PodPending:
				info.PendingPods++
			case corev1.PodFailed:
				info.FailedPods++
			}
		}
	}

	return info, nil
}

// ClusterInfo holds cluster information
type ClusterInfo struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	Connected       bool   `json:"connected"`
	NodesCount      int    `json:"nodes_count"`
	NamespacesCount int    `json:"namespaces_count"`
	PodsCount       int    `json:"pods_count"`
	RunningPods     int    `json:"running_pods"`
	PendingPods     int    `json:"pending_pods"`
	FailedPods      int    `json:"failed_pods"`
	CPUCapacity     string `json:"cpu_capacity"`
	MemoryCapacity  string `json:"memory_capacity"`
}

// GetNamespaces lists all namespaces
func (c *ClusterClient) GetNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	list, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetPods lists pods in a namespace
func (c *ClusterClient) GetPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	list, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetServices lists services in a namespace
func (c *ClusterClient) GetServices(ctx context.Context, namespace string) ([]corev1.Service, error) {
	list, err := c.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetConfigMaps lists configmaps in a namespace
func (c *ClusterClient) GetConfigMaps(ctx context.Context, namespace string) ([]corev1.ConfigMap, error) {
	list, err := c.Clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetSecrets lists secrets in a namespace
func (c *ClusterClient) GetSecrets(ctx context.Context, namespace string) ([]corev1.Secret, error) {
	list, err := c.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetEvents lists events in a namespace
func (c *ClusterClient) GetEvents(ctx context.Context, namespace string) ([]corev1.Event, error) {
	list, err := c.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetPodLogs retrieves logs from a pod
func (c *ClusterClient) GetPodLogs(ctx context.Context, namespace, podName, containerName string, tailLines int64) (string, error) {
	opts := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &tailLines,
	}

	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	logs, err := req.DoRaw(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}

	return string(logs), nil
}
