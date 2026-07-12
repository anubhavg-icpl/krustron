// Krustron Dashboard - Cost API
// Mirrors internal/cost response shapes (all wrapped in {data}).

import api from './client'

export interface CostSummary {
  current_month_cost: number
  previous_month_cost: number
  change_percent: number
  potential_savings: number
  top_namespaces: unknown[]
  currency: string
  generated_at: string
}

export interface CostAllocation {
  id: string
  cluster_id: string
  cluster_name: string
  namespace: string
  workload_type: string
  workload_name: string
  cpu_cost: number
  memory_cost: number
  storage_cost: number
  network_cost: number
  total_cost: number
  efficiency: number
  period_start: string
  period_end: string
}

export interface Budget {
  id: string
  name: string
  type: string
  amount: number
  currency: string
  scope: string
  scope_value: string
  current_spend: number
  forecast_spend: number
  status: string
  period_start: string
  period_end: string
}

export interface AllocationFilter {
  cluster?: string
  namespace?: string
  workload_type?: string
  limit?: number
  signal?: AbortSignal
}

function toQuery(opts: Record<string, unknown>): string {
  const qs = new URLSearchParams(
    Object.entries(opts).filter(([, v]) => v !== undefined && v !== '') as [string, string][]
  ).toString()
  return qs ? `?${qs}` : ''
}

export const costApi = {
  getSummary: async () => {
    const response = await api.get<CostSummary>('/cost/summary')
    return response.data
  },

  listAllocations: async (opts: AllocationFilter = {}) => {
    const { signal, ...rest } = opts
    const response = await api.get<CostAllocation[]>(
      `/cost/allocations${toQuery(rest)}`,
      signal ? { signal } : undefined
    )
    return response.data
  },

  listBudgets: async () => {
    const response = await api.get<Budget[]>('/cost/budgets')
    return response.data
  },

  createBudget: async (data: Partial<Budget>) => {
    const response = await api.post<Budget>('/cost/budgets', data)
    return response.data
  },

  generateReport: async (opts: { type?: string; cluster?: string; namespace?: string } = {}) => {
    const response = await api.post<unknown>(`/cost/reports${toQuery(opts as Record<string, unknown>)}`)
    return response.data
  },
}

export default costApi
