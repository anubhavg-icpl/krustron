// Krustron Dashboard - Type Definitions
// Author: Anubhav Gain <anubhavg@infopercept.com>

// ============================================================================
// WebSocket Message Types
// ============================================================================

export type MessageType =
  | 'cluster.status'
  | 'cluster.health'
  | 'cluster.metrics'
  | 'app.status'
  | 'app.health'
  | 'app.sync'
  | 'pipeline.status'
  | 'pipeline.logs'
  | 'pipeline.stage'
  | 'pod.status'
  | 'pod.logs'
  | 'pod.metrics'
  | 'alert'
  | 'alert.resolved'
  | 'cost.update'
  | 'cost.alert'
  | 'security.scan'
  | 'security.vulnerability'
  | 'ai.response'
  | 'ai.recommendation'
  | 'notification'
  | 'system.status'
  | 'error'
  | 'ping'
  | 'pong'
  | 'subscribe'
  | 'unsubscribe'

export interface WebSocketMessage<T = unknown> {
  id: string
  type: MessageType
  channel: string
  payload: T
  timestamp: string
  metadata?: Record<string, string>
}

// ============================================================================
// Cluster Types
// ============================================================================

export interface Cluster {
  id: string
  name: string
  server: string
  status: ClusterStatus
  version: string
  nodeCount: number
  labels: Record<string, string>
  health: ClusterHealth
  createdAt: string
  updatedAt: string
}

export type ClusterStatus = 'connected' | 'disconnected' | 'unknown'

export interface ClusterHealth {
  status: HealthStatus
  components: ComponentHealth[]
  lastChecked: string
}

export interface ComponentHealth {
  name: string
  status: HealthStatus
  message?: string
}

export type HealthStatus = 'healthy' | 'progressing' | 'degraded' | 'suspended' | 'missing' | 'unknown'

export interface ClusterMetrics {
  clusterId: string
  cpu: ResourceMetrics
  memory: ResourceMetrics
  pods: PodMetrics
  nodes: NodeMetrics[]
  timestamp: string
}

export interface ResourceMetrics {
  used: number
  available: number
  total: number
  percentage: number
}

export interface PodMetrics {
  running: number
  pending: number
  failed: number
  succeeded: number
  total: number
}

export interface NodeMetrics {
  name: string
  status: string
  cpu: ResourceMetrics
  memory: ResourceMetrics
  pods: number
}

// ============================================================================
// Application Types
// ============================================================================

export interface Application {
  id: string
  name: string
  clusterId: string
  clusterName: string
  namespace: string
  project: string
  source: ApplicationSource
  destination: ApplicationDestination
  syncStatus: SyncStatus
  healthStatus: HealthStatus
  operationState?: OperationState
  resources: ApplicationResource[]
  createdAt: string
  updatedAt: string
}

export interface ApplicationSource {
  repoUrl: string
  path: string
  targetRevision: string
  chart?: string
  helm?: HelmSource
  kustomize?: KustomizeSource
}

export interface HelmSource {
  valueFiles?: string[]
  values?: string
  parameters?: HelmParameter[]
  releaseName?: string
}

export interface HelmParameter {
  name: string
  value: string
}

export interface KustomizeSource {
  namePrefix?: string
  nameSuffix?: string
  images?: string[]
  commonLabels?: Record<string, string>
  commonAnnotations?: Record<string, string>
}

export interface ApplicationDestination {
  server: string
  namespace: string
}

export type SyncStatus = 'Synced' | 'OutOfSync' | 'Unknown'

export interface OperationState {
  operation: {
    sync?: {
      revision: string
      prune: boolean
      dryRun: boolean
    }
  }
  phase: OperationPhase
  message: string
  startedAt: string
  finishedAt?: string
}

export type OperationPhase = 'Running' | 'Succeeded' | 'Failed' | 'Error' | 'Terminating'

export interface ApplicationResource {
  group: string
  version: string
  kind: string
  namespace: string
  name: string
  status: SyncStatus
  health?: {
    status: HealthStatus
    message?: string
  }
}

// ============================================================================
// Pipeline Types
// ============================================================================

export interface Pipeline {
  id: string
  name: string
  applicationId?: string
  applicationName?: string
  stages: PipelineStage[]
  triggers: PipelineTrigger[]
  notifications?: PipelineNotifications
  lastRun?: PipelineRun
  runCount: number
  successCount: number
  failureCount: number
  createdAt: string
  updatedAt: string
}

export interface PipelineStage {
  name: string
  type: StageType
  dependsOn?: string[]
  config: StageConfig
  timeout?: string
  retries?: number
}

export type StageType = 'build' | 'test' | 'security' | 'deploy' | 'approval' | 'custom'

export interface StageConfig {
  build?: BuildConfig
  test?: TestConfig
  security?: SecurityConfig
  deploy?: DeployConfig
  approval?: ApprovalConfig
  custom?: CustomConfig
}

export interface BuildConfig {
  image?: string
  dockerfile?: string
  context?: string
  buildArgs?: Record<string, string>
  registry?: string
  tags?: string[]
}

export interface TestConfig {
  image?: string
  command?: string[]
  env?: Record<string, string>
}

export interface SecurityConfig {
  scanType: 'vulnerability' | 'policy' | 'both'
  failOnCritical?: boolean
  failOnHigh?: boolean
}

export interface DeployConfig {
  environment: string
  strategy: 'recreate' | 'rolling' | 'blueGreen' | 'canary'
  canary?: CanaryConfig
}

export interface CanaryConfig {
  steps: CanaryStep[]
}

export interface CanaryStep {
  weight: number
  pause?: {
    duration: string
  }
}

export interface ApprovalConfig {
  approvers: string[]
  minApprovals: number
  timeout: string
}

export interface CustomConfig {
  image: string
  command?: string[]
  script?: string
  env?: Record<string, string>
}

export interface PipelineTrigger {
  type: 'webhook' | 'schedule' | 'manual' | 'gitPush' | 'gitPullRequest'
  webhook?: {
    secretRef: {
      name: string
      key: string
    }
  }
  schedule?: {
    cron: string
    timezone?: string
  }
  git?: {
    branches?: string[]
    paths?: string[]
  }
}

export interface PipelineNotifications {
  slack?: {
    channel: string
    onSuccess?: boolean
    onFailure?: boolean
  }
  email?: {
    recipients: string[]
    onSuccess?: boolean
    onFailure?: boolean
  }
}

export interface PipelineRun {
  id: string
  pipelineId: string
  status: RunStatus
  stages: StageRun[]
  parameters?: Record<string, string>
  triggeredBy: string
  startedAt: string
  finishedAt?: string
}

export type RunStatus = 'Pending' | 'Running' | 'Succeeded' | 'Failed' | 'Cancelled'

export interface StageRun {
  name: string
  status: RunStatus
  startedAt?: string
  finishedAt?: string
  message?: string
  logs?: string
}

// ============================================================================
// Alert Types
// ============================================================================

export interface Alert {
  id: string
  type: AlertType
  severity: AlertSeverity
  title: string
  message: string
  source: string
  labels: Record<string, string>
  annotations?: Record<string, string>
  status: AlertStatus
  startsAt: string
  endsAt?: string
  acknowledgedAt?: string
  acknowledgedBy?: string
  resolvedAt?: string
  resolvedBy?: string
}

export type AlertType = 'cluster' | 'application' | 'pipeline' | 'security' | 'cost' | 'system'
export type AlertSeverity = 'critical' | 'high' | 'medium' | 'low' | 'info'
export type AlertStatus = 'firing' | 'acknowledged' | 'resolved' | 'silenced'

// ============================================================================
// Cost Types
// ============================================================================

export interface CostSummary {
  totalCost: number
  previousPeriodCost: number
  changePercentage: number
  currency: string
  period: {
    start: string
    end: string
  }
  breakdown: CostBreakdown[]
  topClusters: ClusterCost[]
  topNamespaces: NamespaceCost[]
  forecast: CostForecast
}

export interface CostBreakdown {
  category: string
  cost: number
  percentage: number
}

export interface ClusterCost {
  clusterId: string
  clusterName: string
  cost: number
  percentage: number
}

export interface NamespaceCost {
  namespace: string
  clusterId: string
  cost: number
  percentage: number
}

export interface CostForecast {
  estimatedMonthly: number
  estimatedYearly: number
  trend: 'increasing' | 'decreasing' | 'stable'
}

// ============================================================================
// Security Types
// ============================================================================

export interface SecurityScan {
  id: string
  type: 'vulnerability' | 'policy' | 'compliance'
  target: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  results?: SecurityResults
  startedAt: string
  completedAt?: string
}

export interface SecurityResults {
  critical: number
  high: number
  medium: number
  low: number
  informational: number
  vulnerabilities: Vulnerability[]
}

export interface Vulnerability {
  id: string
  cve?: string
  severity: AlertSeverity
  title: string
  description: string
  package: string
  installedVersion: string
  fixedVersion?: string
  references: string[]
}

// ============================================================================
// AI Types
// ============================================================================

export interface AIQuery {
  id: string
  question: string
  context?: Record<string, unknown>
  response?: AIResponse
  createdAt: string
}

export interface AIResponse {
  answer: string
  steps?: string[]
  confidence: number
  relatedDocs?: string[]
  recommendations?: AIRecommendation[]
}

export interface AIRecommendation {
  id: string
  type: 'optimization' | 'security' | 'reliability' | 'cost'
  priority: 'high' | 'medium' | 'low'
  title: string
  description: string
  impact: string
  action: string
  resources: string[]
}

// ============================================================================
// User Types
// ============================================================================

export interface User {
  id: string
  username: string
  email: string
  displayName?: string
  avatar?: string
  roles: string[]
  teams: string[]
  lastLogin?: string
  createdAt: string
}

export interface AuthState {
  isAuthenticated: boolean
  user: User | null
  token: string | null
  refreshToken: string | null
  expiresAt: number | null
}

// ============================================================================
// UI Types
// ============================================================================

export interface BreadcrumbItem {
  label: string
  href?: string
}

export interface TableColumn<T> {
  key: keyof T | string
  label: string
  sortable?: boolean
  render?: (value: T[keyof T], row: T) => React.ReactNode
}

export interface PaginationState {
  page: number
  limit: number
  total: number
  totalPages: number
}

export interface FilterState {
  search?: string
  status?: string[]
  cluster?: string[]
  namespace?: string[]
  severity?: string[]
  dateRange?: {
    start: string
    end: string
  }
}

export interface SortState {
  field: string
  direction: 'asc' | 'desc'
}

// ============================================================================
// Chart Types
// ============================================================================

export interface ChartDataPoint {
  timestamp: string
  value: number
  label?: string
}

export interface TimeSeriesData {
  name: string
  data: ChartDataPoint[]
  color?: string
}

export interface DonutChartData {
  name: string
  value: number
  color: string
}
