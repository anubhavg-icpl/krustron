// Krustron Dashboard - Helm API
// Mirrors internal/helm response shapes (all wrapped in {data}).

import api from './client'

export interface HelmRepository {
  name: string
  url: string
  username?: string
  last_sync: string
  chart_count: number
}

export interface HelmChart {
  name: string
  repository: string
  version: string
  app_version: string
  description: string
  icon: string
  keywords: string[]
  home: string
  sources: string[]
  deprecated: boolean
}

export interface HelmRelease {
  id: string
  name: string
  cluster_id: string
  namespace: string
  chart_name: string
  chart_version: string
  chart_repo: string
  app_version: string
  status: string
  revision: number
  last_deployed: string
  notes: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface AddRepoRequest {
  name: string
  url: string
  username?: string
  password?: string
}

export interface InstallReleaseRequest {
  name: string
  cluster_id: string
  namespace: string
  chart_repo: string
  chart_name: string
  chart_version?: string
  values?: Record<string, unknown>
  values_yaml?: string
  create_namespace?: boolean
  wait?: boolean
  timeout?: number
}

export interface UpgradeReleaseRequest {
  chart_version?: string
  values?: Record<string, unknown>
  values_yaml?: string
  reset_values?: boolean
  reuse_values?: boolean
  wait?: boolean
  timeout?: number
}

export const helmApi = {
  listRepositories: async () => {
    const response = await api.get<HelmRepository[]>('/helm/repositories')
    return response.data
  },
  addRepository: async (data: AddRepoRequest) => {
    const response = await api.post<HelmRepository>('/helm/repositories', data)
    return response.data
  },
  removeRepository: async (name: string) => {
    return api.delete<{ message: string }>(`/helm/repositories/${name}`)
  },
  syncRepository: async (name: string) => {
    const response = await api.post<HelmRepository>(`/helm/repositories/${name}/sync`)
    return response.data
  },
  searchCharts: async (q?: string) => {
    const qs = q ? `?q=${encodeURIComponent(q)}` : ''
    const response = await api.get<HelmChart[]>(`/helm/charts${qs}`)
    return response.data
  },
  getChartDetails: async (repo: string, chart: string) => {
    const response = await api.get<unknown>(`/helm/charts/${repo}/${chart}`)
    return response.data
  },
  getChartVersions: async (repo: string, chart: string) => {
    const response = await api.get<unknown[]>(`/helm/charts/${repo}/${chart}/versions`)
    return response.data
  },
  listReleases: async () => {
    const response = await api.get<HelmRelease[]>('/helm/releases')
    return response.data
  },
  getRelease: async (cluster: string, namespace: string, name: string) => {
    const response = await api.get<HelmRelease>(`/helm/releases/${cluster}/${namespace}/${name}`)
    return response.data
  },
  installRelease: async (data: InstallReleaseRequest) => {
    const response = await api.post<HelmRelease>('/helm/releases', data)
    return response.data
  },
  upgradeRelease: async (cluster: string, namespace: string, name: string, data: UpgradeReleaseRequest) => {
    const response = await api.put<HelmRelease>(`/helm/releases/${cluster}/${namespace}/${name}`, data)
    return response.data
  },
  uninstallRelease: async (cluster: string, namespace: string, name: string) => {
    return api.delete<{ message: string }>(`/helm/releases/${cluster}/${namespace}/${name}`)
  },
  rollbackRelease: async (cluster: string, namespace: string, name: string, revision: number) => {
    const response = await api.post<HelmRelease>(`/helm/releases/${cluster}/${namespace}/${name}/rollback`, { revision })
    return response.data
  },
  getReleaseHistory: async (cluster: string, namespace: string, name: string) => {
    const response = await api.get<unknown[]>(`/helm/releases/${cluster}/${namespace}/${name}/history`)
    return response.data
  },
  getReleaseValues: async (cluster: string, namespace: string, name: string) => {
    const response = await api.get<unknown>(`/helm/releases/${cluster}/${namespace}/${name}/values`)
    return response.data
  },
}

export default helmApi
