// Krustron Dashboard - Alerts API
// Author: Anubhav Gain <anubhavg@infopercept.com>

import api, { RequestConfig } from './client'
import { Alert } from '@/types'

// ============================================================================
// Types
// ============================================================================

export interface ListOptions {
  page?: number
  limit?: number
  signal?: AbortSignal
}

// ============================================================================
// API Functions
// ============================================================================

export const alertsApi = {
  list: async (options: ListOptions = {}) => {
    const { page = 1, limit = 50, signal } = options
    const config: RequestConfig = signal ? { signal } : {}
    const response = await api.get<Alert[]>(`/observability/alerts?page=${page}&limit=${limit}`, config)
    return response.data
  },

  acknowledge: async (id: string) => {
    return api.post<{ message: string }>(`/observability/alerts/${id}/acknowledge`)
  },

  resolve: async (id: string) => {
    return api.post<{ message: string }>(`/observability/alerts/${id}/resolve`)
  },

  silence: async (id: string, duration: string) => {
    return api.post<{ message: string }>(`/observability/alerts/${id}/silence`, { duration })
  },
}

export default alertsApi
