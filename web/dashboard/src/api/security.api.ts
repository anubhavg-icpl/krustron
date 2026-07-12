// Krustron Dashboard - Security API
// Mirrors internal/security response shapes (all wrapped in {data}).

import api from './client'

export interface SecurityScan {
  id: string
  scan_type: string
  target_type: string
  target_id: string
  target_name: string
  cluster_id: string
  status: string
  critical_count: number
  high_count: number
  medium_count: number
  low_count: number
  unknown_count: number
  scanner: string
  scanner_version: string
  started_at?: string
  finished_at?: string
  created_at: string
}

export interface Vulnerability {
  id: string
  vuln_id: string
  package: string
  version: string
  fixed_in: string
  severity: 'critical' | 'high' | 'medium' | 'low' | 'unknown' | string
  title: string
  description: string
  cvss: number
  references: string[]
  cluster_id: string
  namespace: string
  resource: string
}

export interface SecurityPolicy {
  id: string
  name: string
  description: string
  category: string
  severity: string
  rego: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface TriggerScanRequest {
  scan_type: 'container' | 'image' | 'cluster' | 'namespace'
  target_type: string
  target_id: string
  target_name: string
  cluster_id?: string
}

export interface ListOptions {
  page?: number
  limit?: number
  severity?: string
  status?: string
  cluster?: string
  namespace?: string
  target_type?: string
  signal?: AbortSignal
}

function toQuery(opts: Record<string, unknown>): string {
  const qs = new URLSearchParams(
    Object.entries(opts).filter(([, v]) => v !== undefined && v !== '') as [string, string][]
  ).toString()
  return qs ? `?${qs}` : ''
}

export const securityApi = {
  listScans: async (opts: ListOptions = {}) => {
    const { signal, ...rest } = opts
    const response = await api.get<SecurityScan[]>(
      `/security/scans${toQuery(rest)}`,
      signal ? { signal } : undefined
    )
    return response.data
  },

  getScan: async (id: string) => {
    const response = await api.get<SecurityScan>(`/security/scans/${id}`)
    return response.data
  },

  triggerScan: async (data: TriggerScanRequest) => {
    const response = await api.post<SecurityScan>('/security/scans', data)
    return response.data
  },

  listVulnerabilities: async (opts: ListOptions = {}) => {
    const { signal, ...rest } = opts
    const response = await api.get<Vulnerability[]>(
      `/security/vulnerabilities${toQuery(rest)}`,
      signal ? { signal } : undefined
    )
    return response.data
  },

  listPolicies: async () => {
    const response = await api.get<SecurityPolicy[]>('/security/policies')
    return response.data
  },

  createPolicy: async (data: Partial<SecurityPolicy>) => {
    const response = await api.post<SecurityPolicy>('/security/policies', data)
    return response.data
  },

  updatePolicy: async (id: string, data: Partial<SecurityPolicy>) => {
    const response = await api.put<SecurityPolicy>(`/security/policies/${id}`, data)
    return response.data
  },

  deletePolicy: async (id: string) => {
    return api.delete<{ message: string }>(`/security/policies/${id}`)
  },

  validatePolicy: async (id: string) => {
    const response = await api.post<{ valid: boolean; errors?: string[] }>(
      `/security/policies/${id}/validate`
    )
    return response.data
  },
}

export default securityApi
