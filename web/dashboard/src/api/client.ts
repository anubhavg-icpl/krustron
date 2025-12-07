// Krustron Dashboard - API Client
// Author: Anubhav Gain <anubhavg@infopercept.com>
// Centralized API client with interceptors, token refresh, and error handling

import { useAuthStore } from '@/store/useStore'

// ============================================================================
// Types
// ============================================================================

export interface ApiResponse<T = unknown> {
  data: T
  message?: string
}

export interface ApiError {
  code: string
  message: string
  details?: Record<string, unknown>
}

export interface RequestConfig extends RequestInit {
  skipAuth?: boolean
  retries?: number
}

// ============================================================================
// Configuration
// ============================================================================

const API_BASE_URL = '/api/v1'
const MAX_RETRIES = 3
const RETRY_DELAY = 1000

// ============================================================================
// Token Management
// ============================================================================

let isRefreshing = false
let refreshSubscribers: ((token: string) => void)[] = []

const subscribeTokenRefresh = (callback: (token: string) => void) => {
  refreshSubscribers.push(callback)
}

const onTokenRefreshed = (token: string) => {
  refreshSubscribers.forEach((callback) => callback(token))
  refreshSubscribers = []
}

async function refreshToken(): Promise<string | null> {
  const { refreshToken: storedRefreshToken, refreshTokens, logout } = useAuthStore.getState()

  if (!storedRefreshToken) {
    logout()
    return null
  }

  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: storedRefreshToken }),
    })

    if (!response.ok) {
      logout()
      return null
    }

    const data = await response.json()
    const { access_token, refresh_token, expires_in } = data.data

    refreshTokens(access_token, refresh_token, expires_in)
    return access_token
  } catch (error) {
    console.error('[API] Token refresh failed:', error)
    logout()
    return null
  }
}

// ============================================================================
// Request Interceptor
// ============================================================================

async function requestInterceptor(
  url: string,
  config: RequestConfig
): Promise<{ url: string; config: RequestConfig }> {
  const { token, expiresAt } = useAuthStore.getState()

  // Skip auth for public endpoints
  if (config.skipAuth) {
    return { url, config }
  }

  // Check if token needs refresh (refresh 60 seconds before expiry)
  if (token && expiresAt && Date.now() > expiresAt - 60000) {
    if (!isRefreshing) {
      isRefreshing = true
      const newToken = await refreshToken()
      isRefreshing = false

      if (newToken) {
        onTokenRefreshed(newToken)
      }
    }

    // Wait for token refresh if already in progress
    if (isRefreshing) {
      await new Promise<string>((resolve) => {
        subscribeTokenRefresh(resolve)
      })
    }
  }

  // Add auth header
  const currentToken = useAuthStore.getState().token
  if (currentToken) {
    config.headers = {
      ...config.headers,
      Authorization: `Bearer ${currentToken}`,
    }
  }

  return { url, config }
}

// ============================================================================
// Response Interceptor
// ============================================================================

async function responseInterceptor<T>(
  response: Response,
  url: string,
  config: RequestConfig
): Promise<T> {
  // Handle 401 Unauthorized
  if (response.status === 401 && !config.skipAuth) {
    const newToken = await refreshToken()
    if (newToken) {
      // Retry request with new token
      config.headers = {
        ...config.headers,
        Authorization: `Bearer ${newToken}`,
      }
      const retryResponse = await fetch(url, config)
      if (retryResponse.ok) {
        return retryResponse.json()
      }
    }
    throw new ApiClientError('Unauthorized', 'UNAUTHORIZED', 401)
  }

  // Handle other error responses
  if (!response.ok) {
    let errorData: ApiError = { code: 'UNKNOWN', message: 'An error occurred' }
    try {
      const errorJson = await response.json()
      errorData = {
        code: errorJson.code || `HTTP_${response.status}`,
        message: errorJson.message || response.statusText,
        details: errorJson.details,
      }
    } catch {
      errorData.message = response.statusText
    }
    throw new ApiClientError(errorData.message, errorData.code, response.status, errorData.details)
  }

  // Parse successful response
  return response.json()
}

// ============================================================================
// Error Class
// ============================================================================

export class ApiClientError extends Error {
  code: string
  status: number
  details?: Record<string, unknown>

  constructor(
    message: string,
    code: string,
    status: number,
    details?: Record<string, unknown>
  ) {
    super(message)
    this.name = 'ApiClientError'
    this.code = code
    this.status = status
    this.details = details
  }
}

// ============================================================================
// Main Request Function
// ============================================================================

async function request<T>(
  endpoint: string,
  config: RequestConfig = {}
): Promise<ApiResponse<T>> {
  const url = `${API_BASE_URL}${endpoint}`
  const retries = config.retries ?? MAX_RETRIES

  // Default headers
  config.headers = {
    'Content-Type': 'application/json',
    ...config.headers,
  }

  // Apply request interceptor
  const intercepted = await requestInterceptor(url, config)

  let lastError: Error | null = null

  for (let attempt = 0; attempt <= retries; attempt++) {
    try {
      const response = await fetch(intercepted.url, intercepted.config)
      return await responseInterceptor<ApiResponse<T>>(
        response,
        intercepted.url,
        intercepted.config
      )
    } catch (error) {
      lastError = error as Error

      // Don't retry on auth errors or client errors
      if (error instanceof ApiClientError) {
        if (error.status >= 400 && error.status < 500) {
          throw error
        }
      }

      // Wait before retry
      if (attempt < retries) {
        await new Promise((resolve) => setTimeout(resolve, RETRY_DELAY * (attempt + 1)))
        console.warn(`[API] Retrying request (${attempt + 1}/${retries}):`, endpoint)
      }
    }
  }

  throw lastError
}

// ============================================================================
// HTTP Methods
// ============================================================================

export const api = {
  get: <T>(endpoint: string, config?: RequestConfig) =>
    request<T>(endpoint, { ...config, method: 'GET' }),

  post: <T>(endpoint: string, data?: unknown, config?: RequestConfig) =>
    request<T>(endpoint, {
      ...config,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    }),

  put: <T>(endpoint: string, data?: unknown, config?: RequestConfig) =>
    request<T>(endpoint, {
      ...config,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    }),

  patch: <T>(endpoint: string, data?: unknown, config?: RequestConfig) =>
    request<T>(endpoint, {
      ...config,
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined,
    }),

  delete: <T>(endpoint: string, config?: RequestConfig) =>
    request<T>(endpoint, { ...config, method: 'DELETE' }),
}

export default api
