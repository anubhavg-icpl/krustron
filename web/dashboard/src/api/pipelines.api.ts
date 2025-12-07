// Krustron Dashboard - Pipelines API
// Author: Anubhav Gain <anubhavg@infopercept.com>

import api from './client'
import { Pipeline, PipelineRun } from '@/types'

// ============================================================================
// Types
// ============================================================================

export interface CreatePipelineRequest {
  name: string
  description?: string
  git_repository: string
  branch: string
  stages: Array<{
    name: string
    type: string
  }>
  application_id?: string
}

export interface UpdatePipelineRequest {
  name?: string
  description?: string
  stages?: Array<{
    name: string
    type: string
  }>
}

// ============================================================================
// API Functions
// ============================================================================

export const pipelinesApi = {
  list: async (page = 1, limit = 20) => {
    const response = await api.get<Pipeline[]>(`/pipelines?page=${page}&limit=${limit}`)
    return response.data
  },

  get: async (id: string) => {
    const response = await api.get<Pipeline>(`/pipelines/${id}`)
    return response.data
  },

  create: async (data: CreatePipelineRequest) => {
    const response = await api.post<Pipeline>('/pipelines', data)
    return response.data
  },

  update: async (id: string, data: UpdatePipelineRequest) => {
    const response = await api.put<Pipeline>(`/pipelines/${id}`, data)
    return response.data
  },

  delete: async (id: string) => {
    return api.delete<{ message: string }>(`/pipelines/${id}`)
  },

  trigger: async (id: string, parameters?: Record<string, string>) => {
    const response = await api.post<PipelineRun>(`/pipelines/${id}/trigger`, { parameters })
    return response.data
  },

  listRuns: async (id: string, page = 1, limit = 20) => {
    const response = await api.get<PipelineRun[]>(
      `/pipelines/${id}/runs?page=${page}&limit=${limit}`
    )
    return response.data
  },

  getRun: async (id: string, runId: string) => {
    const response = await api.get<PipelineRun>(`/pipelines/${id}/runs/${runId}`)
    return response.data
  },

  cancelRun: async (id: string, runId: string) => {
    return api.post<{ message: string }>(`/pipelines/${id}/runs/${runId}/cancel`)
  },

  retryRun: async (id: string, runId: string) => {
    const response = await api.post<PipelineRun>(`/pipelines/${id}/runs/${runId}/retry`)
    return response.data
  },

  getRunLogs: async (id: string, runId: string) => {
    const response = await api.get<{ logs: string }>(`/pipelines/${id}/runs/${runId}/logs`)
    return response.data
  },
}

export default pipelinesApi
