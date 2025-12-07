# Krustron API Documentation

## Overview

Krustron provides a comprehensive REST API for managing Kubernetes clusters, applications, pipelines, and more. The API follows REST conventions and uses JSON for request/response bodies.

**Base URL:** `https://your-krustron-instance/api/v1`

## Authentication

All API endpoints (except `/auth/login`) require authentication via JWT token.

### Login

```http
POST /auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "your-password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "user": {
    "id": "user-123",
    "username": "admin",
    "email": "admin@example.com",
    "roles": ["super-admin"]
  }
}
```

### Using the Token

Include the token in the `Authorization` header:

```http
GET /api/v1/clusters
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

---

## Clusters

### List Clusters

```http
GET /api/v1/clusters
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20)
- `status` (string): Filter by status (connected, disconnected)

**Response:**
```json
{
  "clusters": [
    {
      "id": "cluster-123",
      "name": "production",
      "server": "https://k8s.example.com:6443",
      "status": "connected",
      "version": "v1.28.0",
      "node_count": 5,
      "labels": {
        "environment": "production"
      },
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

### Get Cluster

```http
GET /api/v1/clusters/{cluster_id}
```

**Response:**
```json
{
  "id": "cluster-123",
  "name": "production",
  "server": "https://k8s.example.com:6443",
  "status": "connected",
  "version": "v1.28.0",
  "node_count": 5,
  "labels": {
    "environment": "production"
  },
  "health": {
    "status": "healthy",
    "components": [
      {"name": "etcd", "status": "healthy"},
      {"name": "scheduler", "status": "healthy"},
      {"name": "controller-manager", "status": "healthy"}
    ]
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Create Cluster

```http
POST /api/v1/clusters
Content-Type: application/json

{
  "name": "staging",
  "server": "https://staging-k8s.example.com:6443",
  "kubeconfig": "base64-encoded-kubeconfig",
  "labels": {
    "environment": "staging"
  }
}
```

### Update Cluster

```http
PUT /api/v1/clusters/{cluster_id}
Content-Type: application/json

{
  "name": "staging-updated",
  "labels": {
    "environment": "staging",
    "team": "platform"
  }
}
```

### Delete Cluster

```http
DELETE /api/v1/clusters/{cluster_id}
```

### Get Cluster Health

```http
GET /api/v1/clusters/{cluster_id}/health
```

### List Cluster Resources

```http
GET /api/v1/clusters/{cluster_id}/resources
```

**Query Parameters:**
- `namespace` (string): Filter by namespace
- `kind` (string): Filter by resource kind (Pod, Deployment, Service)

---

## Applications

### List Applications

```http
GET /api/v1/applications
```

**Query Parameters:**
- `cluster_id` (string): Filter by cluster
- `namespace` (string): Filter by namespace
- `sync_status` (string): Filter by sync status (Synced, OutOfSync)
- `health_status` (string): Filter by health status (Healthy, Degraded)

**Response:**
```json
{
  "applications": [
    {
      "id": "app-123",
      "name": "my-app",
      "cluster_id": "cluster-123",
      "namespace": "default",
      "repo_url": "https://github.com/org/repo",
      "path": "manifests/",
      "target_revision": "main",
      "sync_status": "Synced",
      "health_status": "Healthy",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

### Create Application

```http
POST /api/v1/applications
Content-Type: application/json

{
  "name": "my-app",
  "cluster_id": "cluster-123",
  "namespace": "default",
  "source": {
    "repo_url": "https://github.com/org/repo",
    "path": "manifests/",
    "target_revision": "main"
  },
  "sync_policy": {
    "automated": {
      "prune": true,
      "self_heal": true
    }
  }
}
```

### Sync Application

```http
POST /api/v1/applications/{app_id}/sync
Content-Type: application/json

{
  "prune": true,
  "dry_run": false,
  "revision": "main"
}
```

### Get Application Resources

```http
GET /api/v1/applications/{app_id}/resources
```

### Get Application Events

```http
GET /api/v1/applications/{app_id}/events
```

---

## Pipelines

### List Pipelines

```http
GET /api/v1/pipelines
```

### Create Pipeline

```http
POST /api/v1/pipelines
Content-Type: application/json

{
  "name": "build-deploy",
  "application_id": "app-123",
  "stages": [
    {
      "name": "build",
      "type": "build",
      "config": {
        "dockerfile": "Dockerfile",
        "context": ".",
        "registry": "docker.io/myorg"
      }
    },
    {
      "name": "test",
      "type": "test",
      "depends_on": ["build"],
      "config": {
        "command": ["npm", "test"]
      }
    },
    {
      "name": "deploy",
      "type": "deploy",
      "depends_on": ["test"],
      "config": {
        "environment": "staging",
        "strategy": "rolling"
      }
    }
  ],
  "triggers": [
    {
      "type": "git_push",
      "branches": ["main"]
    }
  ]
}
```

### Trigger Pipeline

```http
POST /api/v1/pipelines/{pipeline_id}/trigger
Content-Type: application/json

{
  "parameters": {
    "IMAGE_TAG": "v1.2.3"
  }
}
```

### Get Pipeline Runs

```http
GET /api/v1/pipelines/{pipeline_id}/runs
```

### Get Pipeline Logs

```http
GET /api/v1/pipelines/{pipeline_id}/runs/{run_id}/logs
```

**Query Parameters:**
- `stage` (string): Filter by stage name
- `follow` (bool): Stream logs in real-time

---

## Helm

### List Repositories

```http
GET /api/v1/helm/repositories
```

### Add Repository

```http
POST /api/v1/helm/repositories
Content-Type: application/json

{
  "name": "bitnami",
  "url": "https://charts.bitnami.com/bitnami"
}
```

### Search Charts

```http
GET /api/v1/helm/charts
```

**Query Parameters:**
- `repo` (string): Repository name
- `keyword` (string): Search keyword

### List Releases

```http
GET /api/v1/helm/releases
```

**Query Parameters:**
- `cluster_id` (string): Filter by cluster
- `namespace` (string): Filter by namespace

### Install Release

```http
POST /api/v1/helm/releases
Content-Type: application/json

{
  "name": "my-nginx",
  "chart": "bitnami/nginx",
  "version": "15.0.0",
  "cluster_id": "cluster-123",
  "namespace": "default",
  "values": {
    "replicaCount": 3,
    "service": {
      "type": "LoadBalancer"
    }
  }
}
```

### Upgrade Release

```http
PUT /api/v1/helm/releases/{release_name}
Content-Type: application/json

{
  "chart": "bitnami/nginx",
  "version": "15.1.0",
  "values": {
    "replicaCount": 5
  }
}
```

### Rollback Release

```http
POST /api/v1/helm/releases/{release_name}/rollback
Content-Type: application/json

{
  "revision": 2
}
```

### Uninstall Release

```http
DELETE /api/v1/helm/releases/{release_name}
```

---

## Security

### Trigger Scan

```http
POST /api/v1/security/scans
Content-Type: application/json

{
  "type": "vulnerability",
  "target": {
    "image": "nginx:1.24"
  }
}
```

### Get Scan Results

```http
GET /api/v1/security/scans/{scan_id}
```

### List Vulnerabilities

```http
GET /api/v1/security/vulnerabilities
```

**Query Parameters:**
- `severity` (string): Filter by severity (CRITICAL, HIGH, MEDIUM, LOW)
- `image` (string): Filter by image

### List Policies

```http
GET /api/v1/security/policies
```

### Create Policy

```http
POST /api/v1/security/policies
Content-Type: application/json

{
  "name": "no-root",
  "description": "Containers must not run as root",
  "rego": "package kubernetes.admission\n\ndeny[msg] {\n  input.request.kind.kind == \"Pod\"\n  container := input.request.object.spec.containers[_]\n  container.securityContext.runAsUser == 0\n  msg := sprintf(\"Container %s runs as root\", [container.name])\n}"
}
```

---

## Observability

### Query Metrics

```http
POST /api/v1/metrics/query
Content-Type: application/json

{
  "query": "sum(rate(container_cpu_usage_seconds_total[5m])) by (pod)",
  "start": "2024-01-15T00:00:00Z",
  "end": "2024-01-15T23:59:59Z",
  "step": "5m"
}
```

### Search Logs

```http
POST /api/v1/logs/search
Content-Type: application/json

{
  "query": "error OR exception",
  "namespace": "default",
  "pod": "my-app-*",
  "start": "2024-01-15T00:00:00Z",
  "end": "2024-01-15T23:59:59Z",
  "limit": 100
}
```

### Get DORA Metrics

```http
GET /api/v1/metrics/dora
```

**Query Parameters:**
- `start` (string): Start date
- `end` (string): End date
- `application` (string): Application name

**Response:**
```json
{
  "deployment_frequency": 4.5,
  "lead_time_for_changes": 2.3,
  "mean_time_to_recovery": 0.5,
  "change_failure_rate": 0.02,
  "period": {
    "start": "2024-01-01",
    "end": "2024-01-31"
  }
}
```

### List Alerts

```http
GET /api/v1/alerts
```

---

## AI Operations

### Ask Question

```http
POST /api/v1/ai/ask
Content-Type: application/json

{
  "question": "Why is my pod CrashLoopBackOff?",
  "context": {
    "pod_name": "my-app-abc123",
    "namespace": "default",
    "cluster_id": "cluster-123"
  }
}
```

**Response:**
```json
{
  "answer": "Based on the pod events and logs...",
  "steps": [
    "Check the container logs with: kubectl logs my-app-abc123",
    "Verify the image exists and is accessible",
    "Check resource limits"
  ],
  "confidence": 0.85,
  "related_docs": [
    "https://kubernetes.io/docs/tasks/debug/..."
  ]
}
```

### Get Recommendations

```http
GET /api/v1/ai/recommendations
```

### Get Chat Sessions

```http
GET /api/v1/ai/sessions
```

---

## Users

### List Users

```http
GET /api/v1/users
```

### Create User

```http
POST /api/v1/users
Content-Type: application/json

{
  "username": "developer",
  "email": "developer@example.com",
  "password": "secure-password",
  "roles": ["developer"]
}
```

### Update User

```http
PUT /api/v1/users/{user_id}
Content-Type: application/json

{
  "email": "newemail@example.com",
  "roles": ["developer", "release-manager"]
}
```

---

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request body",
    "details": [
      {
        "field": "name",
        "message": "name is required"
      }
    ]
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request body |
| `CONFLICT` | 409 | Resource already exists |
| `INTERNAL_ERROR` | 500 | Internal server error |

---

## Rate Limiting

The API implements rate limiting:

- **Default:** 100 requests per minute
- **Authenticated:** 1000 requests per minute

Rate limit headers are included in responses:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1705316400
```

---

## Webhooks

Krustron can send webhooks for various events. Configure webhooks in Settings.

### Event Types

- `cluster.connected`
- `cluster.disconnected`
- `application.synced`
- `application.sync_failed`
- `pipeline.started`
- `pipeline.succeeded`
- `pipeline.failed`
- `security.vulnerability_found`
- `alert.triggered`

### Webhook Payload

```json
{
  "event": "application.synced",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "application_id": "app-123",
    "name": "my-app",
    "revision": "abc123"
  }
}
```

---

## gRPC API

Krustron also provides a gRPC API on port 9090 for high-performance use cases.

```protobuf
service ClusterService {
  rpc ListClusters(ListClustersRequest) returns (ListClustersResponse);
  rpc GetCluster(GetClusterRequest) returns (Cluster);
  rpc WatchClusterEvents(WatchEventsRequest) returns (stream Event);
}
```

See the proto files in `/api/grpc/` for full definitions.
