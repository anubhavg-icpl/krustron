// Krustron Dashboard - Data Fetching Hooks
// Author: Anubhav Gain <anubhavg@infopercept.com>
// Hooks for fetching and syncing data with Zustand stores

import { useEffect, useCallback, useRef } from 'react'
import {
  useAuthStore,
  useClustersStore,
  useApplicationsStore,
  usePipelinesStore,
  useAlertsStore,
} from '@/store/useStore'
import { clustersApi, applicationsApi, pipelinesApi, alertsApi } from '@/api'
import { ApiClientError } from '@/api/client'
import { showErrorToast } from './useNotificationToasts'

// ============================================================================
// Clusters Hook
// ============================================================================

export function useFetchClusters() {
  const { isAuthenticated } = useAuthStore()
  const { setClusters, setLoading, setError } = useClustersStore()
  const hasFetched = useRef(false)
  const abortControllerRef = useRef<AbortController | null>(null)

  const fetchClusters = useCallback(async (signal?: AbortSignal) => {
    if (!isAuthenticated) return

    setLoading(true)
    setError(null)

    try {
      const clusters = await clustersApi.list({ signal })
      // Only update state if request wasn't aborted
      if (!signal?.aborted) {
        setClusters(clusters || [])
      }
    } catch (error) {
      // Ignore abort errors
      if (error instanceof ApiClientError && error.code === 'ABORTED') {
        return
      }
      const message = error instanceof ApiClientError ? error.message : 'Failed to fetch clusters'
      if (!signal?.aborted) {
        setError(message)
        showErrorToast('Failed to load clusters', message)
      }
    } finally {
      if (!signal?.aborted) {
        setLoading(false)
      }
    }
  }, [isAuthenticated, setClusters, setLoading, setError])

  useEffect(() => {
    if (isAuthenticated && !hasFetched.current) {
      hasFetched.current = true
      // Create new AbortController for this fetch
      abortControllerRef.current = new AbortController()
      fetchClusters(abortControllerRef.current.signal)
    }

    // Cleanup: abort pending request on unmount
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort()
        abortControllerRef.current = null
      }
    }
  }, [isAuthenticated, fetchClusters])

  // Reset hasFetched when user logs out
  useEffect(() => {
    if (!isAuthenticated) {
      hasFetched.current = false
    }
  }, [isAuthenticated])

  const refetch = useCallback(() => {
    // Abort any pending request
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }
    // Create new AbortController
    abortControllerRef.current = new AbortController()
    return fetchClusters(abortControllerRef.current.signal)
  }, [fetchClusters])

  return { refetch }
}

// ============================================================================
// Applications Hook
// ============================================================================

export function useFetchApplications() {
  const { isAuthenticated } = useAuthStore()
  const { setApplications, setLoading, setError } = useApplicationsStore()
  const hasFetched = useRef(false)
  const abortControllerRef = useRef<AbortController | null>(null)

  const fetchApplications = useCallback(async (signal?: AbortSignal) => {
    if (!isAuthenticated) return

    setLoading(true)
    setError(null)

    try {
      const applications = await applicationsApi.list({ signal })
      if (!signal?.aborted) {
        setApplications(applications || [])
      }
    } catch (error) {
      if (error instanceof ApiClientError && error.code === 'ABORTED') {
        return
      }
      const message = error instanceof ApiClientError ? error.message : 'Failed to fetch applications'
      if (!signal?.aborted) {
        setError(message)
        showErrorToast('Failed to load applications', message)
      }
    } finally {
      if (!signal?.aborted) {
        setLoading(false)
      }
    }
  }, [isAuthenticated, setApplications, setLoading, setError])

  useEffect(() => {
    if (isAuthenticated && !hasFetched.current) {
      hasFetched.current = true
      abortControllerRef.current = new AbortController()
      fetchApplications(abortControllerRef.current.signal)
    }

    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort()
        abortControllerRef.current = null
      }
    }
  }, [isAuthenticated, fetchApplications])

  useEffect(() => {
    if (!isAuthenticated) {
      hasFetched.current = false
    }
  }, [isAuthenticated])

  const refetch = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }
    abortControllerRef.current = new AbortController()
    return fetchApplications(abortControllerRef.current.signal)
  }, [fetchApplications])

  return { refetch }
}

// ============================================================================
// Pipelines Hook
// ============================================================================

export function useFetchPipelines() {
  const { isAuthenticated } = useAuthStore()
  const { setPipelines, setLoading, setError } = usePipelinesStore()
  const hasFetched = useRef(false)
  const abortControllerRef = useRef<AbortController | null>(null)

  const fetchPipelines = useCallback(async (signal?: AbortSignal) => {
    if (!isAuthenticated) return

    setLoading(true)
    setError(null)

    try {
      const pipelines = await pipelinesApi.list({ signal })
      if (!signal?.aborted) {
        setPipelines(pipelines || [])
      }
    } catch (error) {
      if (error instanceof ApiClientError && error.code === 'ABORTED') {
        return
      }
      const message = error instanceof ApiClientError ? error.message : 'Failed to fetch pipelines'
      if (!signal?.aborted) {
        setError(message)
        showErrorToast('Failed to load pipelines', message)
      }
    } finally {
      if (!signal?.aborted) {
        setLoading(false)
      }
    }
  }, [isAuthenticated, setPipelines, setLoading, setError])

  useEffect(() => {
    if (isAuthenticated && !hasFetched.current) {
      hasFetched.current = true
      abortControllerRef.current = new AbortController()
      fetchPipelines(abortControllerRef.current.signal)
    }

    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort()
        abortControllerRef.current = null
      }
    }
  }, [isAuthenticated, fetchPipelines])

  useEffect(() => {
    if (!isAuthenticated) {
      hasFetched.current = false
    }
  }, [isAuthenticated])

  const refetch = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }
    abortControllerRef.current = new AbortController()
    return fetchPipelines(abortControllerRef.current.signal)
  }, [fetchPipelines])

  return { refetch }
}

// ============================================================================
// Alerts Hook
// ============================================================================

export function useFetchAlerts() {
  const { isAuthenticated } = useAuthStore()
  const { setAlerts, setLoading, setError } = useAlertsStore()
  const hasFetched = useRef(false)
  const abortControllerRef = useRef<AbortController | null>(null)

  const fetchAlerts = useCallback(async (signal?: AbortSignal) => {
    if (!isAuthenticated) return

    setLoading(true)
    setError(null)

    try {
      const alerts = await alertsApi.list({ signal })
      if (!signal?.aborted) {
        setAlerts(alerts || [])
      }
    } catch (error) {
      if (error instanceof ApiClientError && error.code === 'ABORTED') {
        return
      }
      const message = error instanceof ApiClientError ? error.message : 'Failed to fetch alerts'
      if (!signal?.aborted) {
        setError(message)
        // Don't show toast for alerts - not critical
      }
    } finally {
      if (!signal?.aborted) {
        setLoading(false)
      }
    }
  }, [isAuthenticated, setAlerts, setLoading, setError])

  useEffect(() => {
    if (isAuthenticated && !hasFetched.current) {
      hasFetched.current = true
      abortControllerRef.current = new AbortController()
      fetchAlerts(abortControllerRef.current.signal)
    }

    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort()
        abortControllerRef.current = null
      }
    }
  }, [isAuthenticated, fetchAlerts])

  useEffect(() => {
    if (!isAuthenticated) {
      hasFetched.current = false
    }
  }, [isAuthenticated])

  const refetch = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }
    abortControllerRef.current = new AbortController()
    return fetchAlerts(abortControllerRef.current.signal)
  }, [fetchAlerts])

  return { refetch }
}

// ============================================================================
// Combined Data Provider Hook
// ============================================================================

export function useInitializeData() {
  const { refetch: refetchClusters } = useFetchClusters()
  const { refetch: refetchApplications } = useFetchApplications()
  const { refetch: refetchPipelines } = useFetchPipelines()
  const { refetch: refetchAlerts } = useFetchAlerts()

  const refetchAll = useCallback(() => {
    refetchClusters()
    refetchApplications()
    refetchPipelines()
    refetchAlerts()
  }, [refetchClusters, refetchApplications, refetchPipelines, refetchAlerts])

  return { refetchAll }
}

// ============================================================================
// Polling Hook for Real-time Updates Fallback
// ============================================================================

export function useDataPolling(intervalMs = 30000) {
  const { isAuthenticated } = useAuthStore()
  const { refetchAll } = useInitializeData()

  useEffect(() => {
    if (!isAuthenticated) return

    const interval = setInterval(() => {
      refetchAll()
    }, intervalMs)

    return () => clearInterval(interval)
  }, [isAuthenticated, intervalMs, refetchAll])
}
