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
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
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
}

export default authApi
