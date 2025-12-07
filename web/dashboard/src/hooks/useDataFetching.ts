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

  const fetchClusters = useCallback(async () => {
    if (!isAuthenticated) return

    setLoading(true)
    setError(null)

    try {
      const clusters = await clustersApi.list()
      setClusters(clusters || [])
    } catch (error) {
      const message = error instanceof ApiClientError ? error.message : 'Failed to fetch clusters'
      setError(message)
      showErrorToast('Failed to load clusters', message)
    } finally {
      setLoading(false)
    }
  }, [isAuthenticated, setClusters, setLoading, setError])

  useEffect(() => {
    if (isAuthenticated && !hasFetched.current) {
      hasFetched.current = true
      fetchClusters()
    }
  }, [isAuthenticated, fetchClusters])

  return { refetch: fetchClusters }
}

// ============================================================================
// Applications Hook
// ============================================================================

export function useFetchApplications() {
  const { isAuthenticated } = useAuthStore()
  const { setApplications, setLoading, setError } = useApplicationsStore()
  const hasFetched = useRef(false)

  const fetchApplications = useCallback(async () => {
    if (!isAuthenticated) return

    setLoading(true)
    setError(null)

    try {
      const applications = await applicationsApi.list()
      setApplications(applications || [])
    } catch (error) {
      const message = error instanceof ApiClientError ? error.message : 'Failed to fetch applications'
      setError(message)
      showErrorToast('Failed to load applications', message)
    } finally {
      setLoading(false)
    }
  }, [isAuthenticated, setApplications, setLoading, setError])

  useEffect(() => {
    if (isAuthenticated && !hasFetched.current) {
      hasFetched.current = true
      fetchApplications()
    }
  }, [isAuthenticated, fetchApplications])

  return { refetch: fetchApplications }
}

// ============================================================================
// Pipelines Hook
// ============================================================================

export function useFetchPipelines() {
  const { isAuthenticated } = useAuthStore()
  const { setPipelines, setLoading, setError } = usePipelinesStore()
  const hasFetched = useRef(false)

  const fetchPipelines = useCallback(async () => {
    if (!isAuthenticated) return

    setLoading(true)
    setError(null)

    try {
      const pipelines = await pipelinesApi.list()
      setPipelines(pipelines || [])
    } catch (error) {
      const message = error instanceof ApiClientError ? error.message : 'Failed to fetch pipelines'
      setError(message)
      showErrorToast('Failed to load pipelines', message)
    } finally {
      setLoading(false)
    }
  }, [isAuthenticated, setPipelines, setLoading, setError])

  useEffect(() => {
    if (isAuthenticated && !hasFetched.current) {
      hasFetched.current = true
      fetchPipelines()
    }
  }, [isAuthenticated, fetchPipelines])

  return { refetch: fetchPipelines }
}

// ============================================================================
// Alerts Hook
// ============================================================================

export function useFetchAlerts() {
  const { isAuthenticated } = useAuthStore()
  const { setAlerts, setLoading, setError } = useAlertsStore()
  const hasFetched = useRef(false)

  const fetchAlerts = useCallback(async () => {
    if (!isAuthenticated) return

    setLoading(true)
    setError(null)

    try {
      const alerts = await alertsApi.list()
      setAlerts(alerts || [])
    } catch (error) {
      const message = error instanceof ApiClientError ? error.message : 'Failed to fetch alerts'
      setError(message)
      // Don't show toast for alerts - not critical
    } finally {
      setLoading(false)
    }
  }, [isAuthenticated, setAlerts, setLoading, setError])

  useEffect(() => {
    if (isAuthenticated && !hasFetched.current) {
      hasFetched.current = true
      fetchAlerts()
    }
  }, [isAuthenticated, fetchAlerts])

  return { refetch: fetchAlerts }
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
