// Krustron Dashboard - Observability API
// Mirrors internal/observability response shapes (all wrapped in {data}).

import api, { RequestConfig } from './client'

export interface PlatformMetrics {
  clusters: number
  applications: number
  pipelines: number
  deployments_today: number
  success_rate: number
  mttr_hours: number
}

export interface ClusterMetrics {
  cpu_usage: number
  memory_usage: number
  pod_count?: number
  node_count?: number
  [key: string]: unknown
}

export interface MetricsQuery {
  start?: string
  end?: string
  step?: string
}

export interface Alert {
  id: string
  name: string
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info'
  state: 'firing' | 'resolved' | 'pending' | string
  message: string
  cluster_id?: string
  started_at?: string
  ended_at?: string
  [key: string]: unknown
}

export interface AlertFilters {
  state?: string
  severity?: string
  cluster?: string
}

export interface DashboardMeta {
  id: string
  title: string
  url: string
  [key: string]: unknown
}

export const observabilityApi = {
  getMetrics: async (signal?: RequestConfig) => {
    const response = await api.get<PlatformMetrics>('/observability/metrics', signal)
    return response.data
  },

  getClusterMetrics: async (clusterId: string, query: MetricsQuery = {}) => {
    const qs = new URLSearchParams(
      Object.entries(query).filter(([, v]) => v) as [string, string][]
    ).toString()
    const response = await api.get<ClusterMetrics>(`/observability/metrics/clusters/${clusterId}${qs ? `?${qs}` : ''}`)
    return response.data
  },

  getApplicationMetrics: async (applicationId: string, query: MetricsQuery = {}) => {
    const qs = new URLSearchParams(
      Object.entries(query).filter(([, v]) => v) as [string, string][]
    ).toString()
    const response = await api.get<unknown>(`/observability/metrics/applications/${applicationId}${qs ? `?${qs}` : ''}`)
    return response.data
  },

  listAlerts: async (filters: AlertFilters = {}) => {
    const qs = new URLSearchParams(
      Object.entries(filters).filter(([, v]) => v) as [string, string][]
    ).toString()
    const response = await api.get<Alert[]>(`/observability/alerts${qs ? `?${qs}` : ''}`)
    return response.data
  },

  listDashboards: async () => {
    const response = await api.get<DashboardMeta[]>('/observability/dashboards')
    return response.data
  },

  getDora: async (query: { start?: string; end?: string; application?: string; cluster?: string } = {}) => {
    const qs = new URLSearchParams(
      Object.entries(query).filter(([, v]) => v) as [string, string][]
    ).toString()
    const response = await api.get<unknown>(`/observability/dora${qs ? `?${qs}` : ''}`)
    return response.data
  },
}

export default observabilityApi
