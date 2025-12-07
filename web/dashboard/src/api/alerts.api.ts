// Krustron Dashboard - Alerts API
// Author: Anubhav Gain <anubhavg@infopercept.com>

import api from './client'
import { Alert } from '@/types'

// ============================================================================
// API Functions
// ============================================================================

export const alertsApi = {
  list: async (page = 1, limit = 50) => {
    const response = await api.get<Alert[]>(`/observability/alerts?page=${page}&limit=${limit}`)
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
