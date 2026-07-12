// Krustron Dashboard - Auth API
// Author: Anubhav Gain <anubhavg@infopercept.com>

import api from './client'
import { User } from '@/types'

// ============================================================================
// Types
// ============================================================================

export interface LoginRequest {
  email: string
  password: string
  totp_code?: string
}

export interface TwoFactorSetup {
  secret: string
  otpauth_url: string
  qr_code_url: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  user: BackendUser
}

export interface BackendUser {
  id: string
  email: string
  name: string
  avatar_url?: string
  provider: string
  role: string
  is_active: boolean
  totp_enabled?: boolean
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
}

export interface SessionInfo {
  id: string
  created_at: string
  expires_at: string
  current?: boolean
}

// ============================================================================
// User Mapping
// ============================================================================

export function mapBackendUser(user: BackendUser): User {
  return {
    id: user.id,
    username: user.name || user.email.split('@')[0],
    email: user.email,
    displayName: user.name || user.email.split('@')[0],
    avatar: user.avatar_url,
    roles: [user.role],
    teams: [],
    totpEnabled: user.totp_enabled,
    lastLogin: user.last_login_at,
    createdAt: user.created_at,
  }
}

// ============================================================================
// API Functions
// ============================================================================

export const authApi = {
  login: async (data: LoginRequest) => {
    const response = await api.post<LoginResponse>('/auth/login', data, { skipAuth: true })
    return {
      ...response.data,
      user: mapBackendUser(response.data.user),
    }
  },

  register: async (data: RegisterRequest) => {
    const response = await api.post<LoginResponse>('/auth/register', data, { skipAuth: true })
    return {
      ...response.data,
      user: mapBackendUser(response.data.user),
    }
  },

  logout: async () => {
    return api.post<{ message: string }>('/auth/logout')
  },

  getCurrentUser: async () => {
    const response = await api.get<BackendUser>('/auth/me')
    return mapBackendUser(response.data)
  },

  updateProfile: async (data: { name?: string; avatar_url?: string }) => {
    const response = await api.put<BackendUser>('/auth/me', data)
    return mapBackendUser(response.data)
  },

  changePassword: async (data: { current_password: string; new_password: string }) => {
    return api.put<{ message: string }>('/auth/password', data)
  },

  listSessions: async () => {
    const response = await api.get<SessionInfo[]>('/auth/sessions')
    return response.data
  },

  revokeSession: async (id: string) => {
    return api.delete<{ message: string }>(`/auth/sessions/${id}`)
  },

  setup2FA: async () => {
    const response = await api.post<TwoFactorSetup>('/auth/2fa/setup')
    return response.data
  },

  verify2FA: async (code: string) => {
    return api.post<{ message: string }>('/auth/2fa/verify', { code })
  },

  disable2FA: async (code: string) => {
    return api.post<{ message: string }>('/auth/2fa/disable', { code })
  },
}

export default authApi
