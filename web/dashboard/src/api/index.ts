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
