// Krustron Dashboard - Clusters API
// Author: Anubhav Gain <anubhavg@infopercept.com>

import api from './client'
import { Cluster, ClusterHealth, ClusterMetrics } from '@/types'

// ============================================================================
// Types
// ============================================================================

export interface CreateClusterRequest {
  name: string
  server: string
  description?: string
  kubeconfig?: string
  bearer_token?: string
  labels?: Record<string, string>
}

export interface UpdateClusterRequest {
  name?: string
  labels?: Record<string, string>
}

export interface ClusterListResponse {
  data: Cluster[]
  total: number
  page: number
  limit: number
}

// ============================================================================
// API Functions
// ============================================================================

export const clustersApi = {
  list: async (page = 1, limit = 20) => {
    const response = await api.get<Cluster[]>(`/clusters?page=${page}&limit=${limit}`)
    return response.data
  },

  get: async (id: string) => {
    const response = await api.get<Cluster>(`/clusters/${id}`)
    return response.data
  },

  create: async (data: CreateClusterRequest) => {
    const response = await api.post<Cluster>('/clusters', data)
    return response.data
  },

  update: async (id: string, data: UpdateClusterRequest) => {
    const response = await api.put<Cluster>(`/clusters/${id}`, data)
    return response.data
  },

  delete: async (id: string) => {
    return api.delete<{ message: string }>(`/clusters/${id}`)
  },

  getHealth: async (id: string) => {
    const response = await api.get<ClusterHealth>(`/clusters/${id}/health`)
    return response.data
  },

  getMetrics: async (id: string) => {
    const response = await api.get<ClusterMetrics>(`/clusters/${id}/resources`)
    return response.data
  },

  getNamespaces: async (id: string) => {
    const response = await api.get<string[]>(`/clusters/${id}/namespaces`)
    return response.data
  },

  installAgent: async (id: string) => {
    const response = await api.post<{ manifest: string }>(`/clusters/${id}/agent/install`)
    return response.data
  },
}

export default clustersApi
