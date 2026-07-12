// Krustron Dashboard - API Index
// Author: Anubhav Gain <anubhavg@infopercept.com>

export { default as api, ApiClientError } from './client'
export type { ApiResponse, ApiError, RequestConfig } from './client'

export { default as authApi, mapBackendUser } from './auth.api'
export type { LoginRequest, LoginResponse, BackendUser, RegisterRequest } from './auth.api'

export { default as clustersApi } from './clusters.api'
export type { CreateClusterRequest, UpdateClusterRequest } from './clusters.api'

export { default as applicationsApi } from './applications.api'
export type { CreateApplicationRequest, UpdateApplicationRequest } from './applications.api'

export { default as pipelinesApi } from './pipelines.api'
export type { CreatePipelineRequest, UpdatePipelineRequest } from './pipelines.api'

export { default as alertsApi } from './alerts.api'

export { default as observabilityApi } from './observability.api'
export type { PlatformMetrics, ClusterMetrics, Alert as ObservabilityAlert } from './observability.api'

export { default as securityApi } from './security.api'
export type { SecurityScan, Vulnerability, SecurityPolicy } from './security.api'

export { default as helmApi } from './helm.api'
export type { HelmRepository, HelmChart, HelmRelease } from './helm.api'

export { rbacApi, auditApi, settingsApi } from './rbac.api'
export type { Role, AuditLog, AppSettings, NotificationSettings } from './rbac.api'
