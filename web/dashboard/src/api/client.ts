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
  signal?: AbortSignal
}

// ============================================================================
// Configuration
// ============================================================================

const API_BASE_URL = '/api/v1'
const MAX_RETRIES = 3
const INITIAL_RETRY_DELAY = 1000
const MAX_RETRY_DELAY = 30000

// ============================================================================
// Circuit Breaker
// ============================================================================

interface CircuitBreakerState {
  failures: number
  lastFailure: number
  isOpen: boolean
}

const circuitBreaker: CircuitBreakerState = {
  failures: 0,
  lastFailure: 0,
  isOpen: false,
}

const CIRCUIT_BREAKER_THRESHOLD = 5
const CIRCUIT_BREAKER_RESET_TIMEOUT = 60000 // 1 minute

function checkCircuitBreaker(): boolean {
  if (!circuitBreaker.isOpen) return true

  // Check if enough time has passed to try again
  if (Date.now() - circuitBreaker.lastFailure > CIRCUIT_BREAKER_RESET_TIMEOUT) {
    circuitBreaker.isOpen = false
    circuitBreaker.failures = 0
    return true
  }

  return false
}

function recordFailure(): void {
  circuitBreaker.failures++
  circuitBreaker.lastFailure = Date.now()

  if (circuitBreaker.failures >= CIRCUIT_BREAKER_THRESHOLD) {
    circuitBreaker.isOpen = true
    console.warn('[API] Circuit breaker opened - too many failures')
  }
}

function recordSuccess(): void {
  circuitBreaker.failures = 0
  circuitBreaker.isOpen = false
}

// ============================================================================
// Token Management with Proper Mutex
// ============================================================================

let refreshPromise: Promise<string | null> | null = null

async function getValidToken(): Promise<string | null> {
  const { token, expiresAt } = useAuthStore.getState()

  // Token doesn't need refresh
  if (token && expiresAt && Date.now() < expiresAt - 60000) {
    return token
  }

  // Token needs refresh - use mutex pattern
  if (!refreshPromise) {
    refreshPromise = refreshToken().finally(() => {
      refreshPromise = null
    })
  }

  return refreshPromise
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
  // Skip auth for public endpoints
  if (config.skipAuth) {
    return { url, config }
  }

  // Get valid token (handles refresh with proper mutex)
  const currentToken = await getValidToken()

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
// Exponential Backoff Calculator
// ============================================================================

function calculateBackoff(attempt: number): number {
  // Exponential backoff with jitter: min(maxDelay, baseDelay * 2^attempt) + random jitter
  const exponentialDelay = INITIAL_RETRY_DELAY * Math.pow(2, attempt)
  const jitter = Math.random() * 1000
  return Math.min(MAX_RETRY_DELAY, exponentialDelay + jitter)
}

// ============================================================================
// Main Request Function
// ============================================================================

async function request<T>(
  endpoint: string,
  config: RequestConfig = {}
): Promise<ApiResponse<T>> {
  // Check circuit breaker before making request
  if (!checkCircuitBreaker()) {
    throw new ApiClientError(
      'Service temporarily unavailable - too many failures',
      'CIRCUIT_BREAKER_OPEN',
      503
    )
  }

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
      // Check if request was aborted
      if (config.signal?.aborted) {
        throw new ApiClientError('Request aborted', 'ABORTED', 0)
      }

      const response = await fetch(intercepted.url, {
        ...intercepted.config,
        signal: config.signal,
      })

      const result = await responseInterceptor<ApiResponse<T>>(
        response,
        intercepted.url,
        intercepted.config
      )

      // Record success for circuit breaker
      recordSuccess()
      return result
    } catch (error) {
      lastError = error as Error

      // Don't retry on aborted requests
      if (error instanceof DOMException && error.name === 'AbortError') {
        throw new ApiClientError('Request aborted', 'ABORTED', 0)
      }

      // Don't retry on auth errors or client errors
      if (error instanceof ApiClientError) {
        if (error.status >= 400 && error.status < 500) {
          throw error
        }
        // Record failure for server errors
        if (error.status >= 500) {
          recordFailure()
        }
      }

      // Wait before retry with exponential backoff
      if (attempt < retries) {
        const backoffDelay = calculateBackoff(attempt)
        console.warn(`[API] Retrying request in ${backoffDelay}ms (attempt ${attempt + 1}/${retries}):`, endpoint)
        await new Promise((resolve) => setTimeout(resolve, backoffDelay))
      }
    }
  }

  // All retries exhausted - record failure
  recordFailure()
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
