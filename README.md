# Krustron

**Kubernetes-Native Platform for CI/CD, GitOps, and Cluster Management**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28+-326CE5?style=flat&logo=kubernetes)](https://kubernetes.io/)

Krustron is an open-source Devtron alternative that combines a unified dashboard, end-to-end GitOps CI/CD pipelines, multi-cluster management, and integrated observability/security with fine-grained RBAC and AI-assisted operations.

## Author

**Anubhav Gain** - <anubhavg@infopercept.com>

## Features

### Core Platform
- **Multi-Cluster Management** - Manage 50+ Kubernetes clusters from a single pane of glass
- **Unified Dashboard** - Real-time resource browser for Nodes, Pods, Deployments, and CRDs
- **Fine-grained RBAC** - Role-based access control with OIDC/SSO integration

### GitOps & CI/CD
- **Visual Pipeline Builder** - Drag-and-drop YAML editor for build, deploy, and promote stages
- **ArgoCD Integration** - Native GitOps with automatic sync, prune, and self-heal
- **Canary & Blue/Green Deployments** - Traffic splitting via Istio/Linkerd
- **DORA Metrics** - Built-in deployment frequency, lead time, and MTTR tracking

### Helm Management
- **Repository Management** - Add, sync, and search Helm repositories
- **Release Management** - Install, upgrade, rollback, and uninstall releases
- **Drift Detection** - Alert when cluster state differs from Helm values
- **History & Rollback** - Full release history with one-click rollback

### Security
- **Container Scanning** - Trivy integration for vulnerability detection
- **Policy Enforcement** - OPA policies with pre-deploy validation
- **Wazuh Integration** - Host-level security monitoring
- **Security Gates** - Block deployments on critical vulnerabilities

### Observability
- **Metrics** - Prometheus integration with Grafana dashboards
- **Logging** - Centralized logging with OpenSearch/Elasticsearch
- **Tracing** - Distributed tracing with Jaeger/OpenTelemetry
- **Alerting** - Alertmanager integration with multi-channel notifications

### AI Operations (Coming Soon)
- **Natural Language Queries** - "Why is payment-service crashing?"
- **Auto-remediation** - Event-driven automation for common issues
- **Cost Optimization** - AI-powered rightsizing recommendations

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Krustron Platform                        │
├─────────────┬─────────────┬─────────────┬─────────────┬─────────┤
│   API       │   GitOps    │  Pipeline   │  Security   │   AI    │
│   Server    │   Engine    │   Engine    │   Scanner   │   Ops   │
├─────────────┴─────────────┴─────────────┴─────────────┴─────────┤
│                    Core Services Layer                           │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐            │
│  │ Cluster  │ │   Helm   │ │   Auth   │ │Observability│         │
│  │ Manager  │ │ Manager  │ │  (RBAC)  │ │  (OTEL)  │           │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘            │
├─────────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐            │
│  │PostgreSQL│ │  Redis   │ │   NATS   │ │  ArgoCD  │            │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘            │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
        ┌──────────────────────────────────────────┐
        │            Kubernetes Clusters            │
        │  ┌────────┐  ┌────────┐  ┌────────┐     │
        │  │ Prod   │  │ Stage  │  │  Dev   │     │
        │  │Cluster │  │Cluster │  │Cluster │     │
        │  └────────┘  └────────┘  └────────┘     │
        └──────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Kubernetes cluster (local or remote)
- PostgreSQL 14+
- Redis 7+

### Local Development

```bash
# Clone the repository
git clone https://github.com/anubhavg-icpl/krustron.git
cd krustron

# Install dependencies
make deps

# Run database migrations
make migrate-up

# Start the server
make run

# Or use hot reload
make dev
```

### Docker

```bash
# Build the image
docker build -t krustron:latest .

# Run with Docker Compose
docker-compose up -d
```

### Kubernetes (Helm)

```bash
# Add the Helm repository
helm repo add krustron https://charts.krustron.io

# Install Krustron
helm install krustron krustron/krustron \
  --namespace krustron \
  --create-namespace \
  --set postgresql.enabled=true \
  --set redis.enabled=true
```

## Configuration

Krustron can be configured via:
- Configuration file (`config.yaml`)
- Environment variables (prefixed with `KRUSTRON_`)
- Command-line flags

See [config.yaml](config.yaml) for all available options.

### Environment Variables

```bash
# Database
export KRUSTRON_DATABASE_PASSWORD=your-password

# Authentication
export KRUSTRON_AUTH_JWT_SECRET=your-jwt-secret

# OIDC (optional)
export KRUSTRON_AUTH_OIDC_CLIENT_SECRET=your-oidc-secret

# ArgoCD (optional)
export KRUSTRON_GITOPS_ARGOCD_AUTH_TOKEN=your-argocd-token
```

## API Documentation

API documentation is available at `/swagger/index.html` when running the server.

### Key Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/clusters` | List all clusters |
| `POST /api/v1/clusters` | Add a new cluster |
| `GET /api/v1/applications` | List GitOps applications |
| `POST /api/v1/applications/:id/sync` | Trigger application sync |
| `GET /api/v1/helm/releases` | List Helm releases |
| `POST /api/v1/pipelines/:id/trigger` | Trigger pipeline |
| `GET /api/v1/security/scans` | List security scans |

## Development

### Project Structure

```
krustron/
├── cmd/krustron/         # Main entry point
├── api/
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   └── router/           # Route definitions
├── internal/
│   ├── cluster/          # Cluster management
│   ├── helm/             # Helm operations
│   ├── gitops/           # GitOps (ArgoCD)
│   ├── pipeline/         # CI/CD pipelines
│   ├── auth/             # Authentication & RBAC
│   ├── security/         # Security scanning
│   └── observability/    # Metrics, logs, traces
├── pkg/
│   ├── kube/             # Kubernetes client
│   ├── logger/           # Logging
│   ├── config/           # Configuration
│   ├── errors/           # Error handling
│   ├── database/         # Database utilities
│   └── cache/            # Redis cache
├── charts/krustron/      # Helm chart
└── docs/                 # Documentation
```

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run with coverage
make test-coverage
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build Docker image
make docker-build
```

## Roadmap

### Phase 1: MVP Dashboard (Weeks 1-4)
- [x] Resource browser
- [x] Helm app management
- [x] Basic RBAC
- [x] Multi-cluster support

### Phase 2: CI/CD GitOps (Weeks 5-10)
- [x] Visual pipeline builder
- [x] ArgoCD integration
- [ ] Canary deployments
- [x] DORA metrics

### Phase 3: Enterprise Extensions (Weeks 11-18)
- [ ] AI operations (LLM integration)
- [ ] Cost management (KubeCost)
- [ ] Auto-remediation
- [ ] 100+ integrations

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [docs.krustron.io](https://docs.krustron.io)
- **Issues**: [GitHub Issues](https://github.com/anubhavg-icpl/krustron/issues)
- **Discussions**: [GitHub Discussions](https://github.com/anubhavg-icpl/krustron/discussions)

## Acknowledgments

- Inspired by [Devtron](https://github.com/devtron-labs/devtron)
- Built with [Gin](https://github.com/gin-gonic/gin), [client-go](https://github.com/kubernetes/client-go), and [ArgoCD](https://github.com/argoproj/argo-cd)

---

**Made with love by [Anubhav Gain](mailto:anubhavg@infopercept.com)**
