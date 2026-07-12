// Krustron Dashboard - RBAC + Audit + Settings API (admin endpoints).
// Mirrors internal/auth Role / AuditLog response shapes (all wrapped in {data}).

import api from './client'

export interface Role {
  id: string
  name: string
  display_name: string
  description: string
  permissions: string[]
  is_system: boolean
  created_at: string
  updated_at: string
}

export interface AuditLog {
  id: string
  user_id?: string
  user_email: string
  action: string
  resource_type: string
  resource_id: string
  resource_name: string
  cluster_id?: string
  cluster_name: string
  metadata?: Record<string, unknown>
  ip_address: string
  status: string
  created_at: string
}

export interface AppSettings {
  auth?: { oidc_enabled?: boolean; session_timeout?: string | number }
  [key: string]: unknown
}

export interface NotificationSettings {
  email?: { enabled?: boolean; [k: string]: unknown }
  slack?: { enabled?: boolean; [k: string]: unknown }
  webhook?: { enabled?: boolean; [k: string]: unknown }
  [key: string]: unknown
}

function toQuery(opts: Record<string, unknown>): string {
  const qs = new URLSearchParams(
    Object.entries(opts).filter(([, v]) => v !== undefined && v !== '') as [string, string][]
  ).toString()
  return qs ? `?${qs}` : ''
}

export const rbacApi = {
  listRoles: async () => {
    const response = await api.get<Role[]>('/rbac/roles')
    return response.data
  },
  getRole: async (id: string) => {
    const response = await api.get<Role>(`/rbac/roles/${id}`)
    return response.data
  },
  createRole: async (data: { name: string; display_name?: string; description?: string; permissions: string[] }) => {
    const response = await api.post<Role>('/rbac/roles', data)
    return response.data
  },
  updateRole: async (id: string, data: Partial<Role>) => {
    const response = await api.put<Role>(`/rbac/roles/${id}`, data)
    return response.data
  },
  deleteRole: async (id: string) => {
    return api.delete<{ message: string }>(`/rbac/roles/${id}`)
  },
  listPermissions: async () => {
    const response = await api.get<string[]>('/rbac/permissions')
    return response.data
  },
}

export const auditApi = {
  listLogs: async (opts: { page?: number; limit?: number; signal?: AbortSignal } = {}) => {
    const { signal, ...rest } = opts
    const response = await api.get<AuditLog[]>(`/audit/logs${toQuery(rest)}`, signal ? { signal } : undefined)
    return response.data
  },
  getLog: async (id: string) => {
    const response = await api.get<AuditLog>(`/audit/logs/${id}`)
    return response.data
  },
  exportLogs: async (format: 'json' | 'csv' = 'json') => {
    // Returns a raw blob; the envelope is bypassed for file downloads.
    const response = await api.get<unknown>(`/audit/logs/export?format=${format}`)
    return response.data
  },
}

export const settingsApi = {
  get: async () => {
    const response = await api.get<AppSettings>('/settings')
    return response.data
  },
  update: async (data: AppSettings) => {
    const response = await api.put<AppSettings>('/settings', data)
    return response.data
  },
  getNotifications: async () => {
    const response = await api.get<NotificationSettings>('/settings/notifications')
    return response.data
  },
  updateNotifications: async (data: NotificationSettings) => {
    const response = await api.put<NotificationSettings>('/settings/notifications', data)
    return response.data
  },
}

export default { rbacApi, auditApi, settingsApi }
