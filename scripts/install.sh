#!/bin/bash
# Krustron Installation Script
# Author: Anubhav Gain <anubhavg@infopercept.com>

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
KRUSTRON_NAMESPACE="krustron"
KRUSTRON_VERSION="latest"
HELM_RELEASE_NAME="krustron"
INSTALL_TYPE="helm"
POSTGRES_ENABLED="true"
REDIS_ENABLED="true"
INGRESS_ENABLED="false"
DRY_RUN="false"

# Print banner
print_banner() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║                                                               ║"
    echo "║   ██╗  ██╗██████╗ ██╗   ██╗███████╗████████╗██████╗  ██████╗ ███╗   ██╗ ║"
    echo "║   ██║ ██╔╝██╔══██╗██║   ██║██╔════╝╚══██╔══╝██╔══██╗██╔═══██╗████╗  ██║ ║"
    echo "║   █████╔╝ ██████╔╝██║   ██║███████╗   ██║   ██████╔╝██║   ██║██╔██╗ ██║ ║"
    echo "║   ██╔═██╗ ██╔══██╗██║   ██║╚════██║   ██║   ██╔══██╗██║   ██║██║╚██╗██║ ║"
    echo "║   ██║  ██╗██║  ██║╚██████╔╝███████║   ██║   ██║  ██║╚██████╔╝██║ ╚████║ ║"
    echo "║   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ║"
    echo "║                                                               ║"
    echo "║          Kubernetes-Native Platform for CI/CD & GitOps        ║"
    echo "║                                                               ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -n, --namespace       Kubernetes namespace (default: krustron)"
    echo "  -v, --version         Krustron version (default: latest)"
    echo "  -r, --release-name    Helm release name (default: krustron)"
    echo "  -t, --type            Installation type: helm, manifest (default: helm)"
    echo "      --no-postgres     Disable PostgreSQL installation"
    echo "      --no-redis        Disable Redis installation"
    echo "      --ingress         Enable ingress"
    echo "      --dry-run         Print commands without executing"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Install with defaults"
    echo "  $0 -n my-namespace -v 1.0.0           # Install specific version"
    echo "  $0 --ingress --no-postgres            # With ingress, without postgres"
}

# Log functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed"
        exit 1
    fi
    log_info "kubectl found: $(kubectl version --client --short 2>/dev/null || echo 'version check failed')"

    # Check helm
    if [ "$INSTALL_TYPE" = "helm" ]; then
        if ! command -v helm &> /dev/null; then
            log_error "helm is not installed"
            exit 1
        fi
        log_info "helm found: $(helm version --short)"
    fi

    # Check cluster connection
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    log_info "Connected to Kubernetes cluster"

    # Check cluster version
    KUBE_VERSION=$(kubectl version --short 2>/dev/null | grep Server | awk '{print $3}' || echo "unknown")
    log_info "Kubernetes version: $KUBE_VERSION"
}

# Create namespace
create_namespace() {
    log_info "Creating namespace: $KRUSTRON_NAMESPACE"

    if [ "$DRY_RUN" = "true" ]; then
        echo "kubectl create namespace $KRUSTRON_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -"
        return
    fi

    kubectl create namespace "$KRUSTRON_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
}

# Install with Helm
install_helm() {
    log_info "Installing Krustron with Helm..."

    # Add Helm repository
    log_info "Adding Helm repository..."
    if [ "$DRY_RUN" = "true" ]; then
        echo "helm repo add krustron https://charts.krustron.io"
        echo "helm repo update"
    else
        helm repo add krustron https://charts.krustron.io 2>/dev/null || true
        helm repo update
    fi

    # Build Helm values
    HELM_VALUES=""
    HELM_VALUES="$HELM_VALUES --set postgresql.enabled=$POSTGRES_ENABLED"
    HELM_VALUES="$HELM_VALUES --set redis.enabled=$REDIS_ENABLED"
    HELM_VALUES="$HELM_VALUES --set ingress.enabled=$INGRESS_ENABLED"

    # Install
    log_info "Installing Helm chart..."
    if [ "$DRY_RUN" = "true" ]; then
        echo "helm upgrade --install $HELM_RELEASE_NAME krustron/krustron \\"
        echo "  --namespace $KRUSTRON_NAMESPACE \\"
        echo "  --version $KRUSTRON_VERSION \\"
        echo "  $HELM_VALUES"
    else
        helm upgrade --install "$HELM_RELEASE_NAME" krustron/krustron \
            --namespace "$KRUSTRON_NAMESPACE" \
            --version "$KRUSTRON_VERSION" \
            $HELM_VALUES \
            --wait \
            --timeout 10m
    fi
}

# Install with manifests
install_manifest() {
    log_info "Installing Krustron with manifests..."

    MANIFEST_URL="https://raw.githubusercontent.com/anubhavg-icpl/krustron/$KRUSTRON_VERSION/manifests/install.yaml"

    if [ "$DRY_RUN" = "true" ]; then
        echo "kubectl apply -f $MANIFEST_URL"
    else
        kubectl apply -f "$MANIFEST_URL"
    fi
}

# Wait for pods
wait_for_pods() {
    log_info "Waiting for pods to be ready..."

    if [ "$DRY_RUN" = "true" ]; then
        echo "kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=krustron -n $KRUSTRON_NAMESPACE --timeout=300s"
        return
    fi

    kubectl wait --for=condition=ready pod \
        -l app.kubernetes.io/name=krustron \
        -n "$KRUSTRON_NAMESPACE" \
        --timeout=300s
}

# Print access information
print_access_info() {
    echo ""
    log_info "Installation complete!"
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                    Krustron Installed!                         ║${NC}"
    echo -e "${GREEN}╠════════════════════════════════════════════════════════════════╣${NC}"
    echo -e "${GREEN}║${NC} Namespace:     $KRUSTRON_NAMESPACE"
    echo -e "${GREEN}║${NC} Release:       $HELM_RELEASE_NAME"
    echo -e "${GREEN}║${NC} Version:       $KRUSTRON_VERSION"
    echo -e "${GREEN}╠════════════════════════════════════════════════════════════════╣${NC}"
    echo -e "${GREEN}║${NC} Access the dashboard:"
    echo -e "${GREEN}║${NC}"
    echo -e "${GREEN}║${NC}   kubectl port-forward svc/$HELM_RELEASE_NAME 8080:80 -n $KRUSTRON_NAMESPACE"
    echo -e "${GREEN}║${NC}   Open: http://localhost:8080"
    echo -e "${GREEN}║${NC}"
    echo -e "${GREEN}║${NC} Default credentials:"
    echo -e "${GREEN}║${NC}   Username: admin"
    echo -e "${GREEN}║${NC}   Password: (run command below)"
    echo -e "${GREEN}║${NC}"
    echo -e "${GREEN}║${NC}   kubectl get secret $HELM_RELEASE_NAME -n $KRUSTRON_NAMESPACE -o jsonpath='{.data.admin-password}' | base64 -d"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -n|--namespace)
                KRUSTRON_NAMESPACE="$2"
                shift 2
                ;;
            -v|--version)
                KRUSTRON_VERSION="$2"
                shift 2
                ;;
            -r|--release-name)
                HELM_RELEASE_NAME="$2"
                shift 2
                ;;
            -t|--type)
                INSTALL_TYPE="$2"
                shift 2
                ;;
            --no-postgres)
                POSTGRES_ENABLED="false"
                shift
                ;;
            --no-redis)
                REDIS_ENABLED="false"
                shift
                ;;
            --ingress)
                INGRESS_ENABLED="true"
                shift
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Main function
main() {
    parse_args "$@"
    print_banner
    check_prerequisites
    create_namespace

    case $INSTALL_TYPE in
        helm)
            install_helm
            ;;
        manifest)
            install_manifest
            ;;
        *)
            log_error "Unknown installation type: $INSTALL_TYPE"
            exit 1
            ;;
    esac

    if [ "$DRY_RUN" != "true" ]; then
        wait_for_pods
    fi

    print_access_info
}

main "$@"
