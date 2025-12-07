// Krustron Dashboard - Applications API
// Author: Anubhav Gain <anubhavg@infopercept.com>

import api, { RequestConfig } from './client'
import { Application, ApplicationResource } from '@/types'

// ============================================================================
// Types
// ============================================================================

export interface CreateApplicationRequest {
  name: string
  cluster_id: string
  namespace: string
  project?: string
  source: {
    repo_url: string
    path: string
    target_revision: string
  }
  sync_policy?: {
    automated?: {
      prune?: boolean
      self_heal?: boolean
    }
  }
}

export interface UpdateApplicationRequest {
  source?: {
    repo_url?: string
    path?: string
    target_revision?: string
  }
  sync_policy?: {
    automated?: {
      prune?: boolean
      self_heal?: boolean
    }
  }
}

export interface ListOptions {
  page?: number
  limit?: number
  signal?: AbortSignal
}

// ============================================================================
// API Functions
// ============================================================================

export const applicationsApi = {
  list: async (options: ListOptions = {}) => {
    const { page = 1, limit = 20, signal } = options
    const config: RequestConfig = signal ? { signal } : {}
    const response = await api.get<Application[]>(`/applications?page=${page}&limit=${limit}`, config)
    return response.data
  },

  get: async (id: string) => {
    const response = await api.get<Application>(`/applications/${id}`)
    return response.data
  },

  create: async (data: CreateApplicationRequest) => {
    const response = await api.post<Application>('/applications', data)
    return response.data
  },

  update: async (id: string, data: UpdateApplicationRequest) => {
    const response = await api.put<Application>(`/applications/${id}`, data)
    return response.data
  },

  delete: async (id: string) => {
    return api.delete<{ message: string }>(`/applications/${id}`)
  },

  sync: async (id: string, options?: { prune?: boolean; dry_run?: boolean }) => {
    const response = await api.post<{ message: string }>(`/applications/${id}/sync`, options)
    return response.data
  },

  getStatus: async (id: string) => {
    const response = await api.get<{ sync_status: string; health_status: string }>(
      `/applications/${id}/status`
    )
    return response.data
  },

  getResources: async (id: string) => {
    const response = await api.get<ApplicationResource[]>(`/applications/${id}/resources`)
    return response.data
  },

  getManifests: async (id: string) => {
    const response = await api.get<{ manifests: string }>(`/applications/${id}/manifests`)
    return response.data
  },

  getDiff: async (id: string) => {
    const response = await api.get<{ diff: string }>(`/applications/${id}/diff`)
    return response.data
  },
}

export default applicationsApi
