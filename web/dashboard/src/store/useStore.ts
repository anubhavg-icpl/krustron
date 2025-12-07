// Krustron Dashboard - Global State Store
// Author: Anubhav Gain <anubhavg@infopercept.com>
// Using Zustand for state management

import { create } from 'zustand'
import { persist, devtools } from 'zustand/middleware'
import {
  Cluster,
  Application,
  Pipeline,
  Alert,
  User,
  AuthState,
  CostSummary,
  FilterState,
  SortState,
  PaginationState,
} from '@/types'

// ============================================================================
// Auth Store
// ============================================================================

interface AuthStore extends AuthState {
  login: (token: string, refreshToken: string, user: User, expiresIn: number) => void
  logout: () => void
  updateUser: (user: Partial<User>) => void
  refreshTokens: (token: string, refreshToken: string, expiresIn: number) => void
}

export const useAuthStore = create<AuthStore>()(
  persist(
    devtools(
      (set) => ({
        isAuthenticated: false,
        user: null,
        token: null,
        refreshToken: null,
        expiresAt: null,

        login: (token, refreshToken, user, expiresIn) => {
          set({
            isAuthenticated: true,
            token,
            refreshToken,
            user,
            expiresAt: Date.now() + expiresIn * 1000,
          })
        },

        logout: () => {
          set({
            isAuthenticated: false,
            token: null,
            refreshToken: null,
            user: null,
            expiresAt: null,
          })
        },

        updateUser: (userData) => {
          set((state) => ({
            user: state.user ? { ...state.user, ...userData } : null,
          }))
        },

        refreshTokens: (token, refreshToken, expiresIn) => {
          set({
            token,
            refreshToken,
            expiresAt: Date.now() + expiresIn * 1000,
          })
        },
      }),
      { name: 'auth-store' }
    ),
    {
      name: 'krustron-auth',
      partialize: (state) => ({
        token: state.token,
        refreshToken: state.refreshToken,
        user: state.user,
        expiresAt: state.expiresAt,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)

// ============================================================================
// Clusters Store
// ============================================================================

interface ClustersStore {
  clusters: Cluster[]
  selectedCluster: Cluster | null
  loading: boolean
  error: string | null
  filter: FilterState
  sort: SortState
  pagination: PaginationState

  setClusters: (clusters: Cluster[]) => void
  addCluster: (cluster: Cluster) => void
  updateCluster: (id: string, updates: Partial<Cluster>) => void
  removeCluster: (id: string) => void
  selectCluster: (cluster: Cluster | null) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  setFilter: (filter: Partial<FilterState>) => void
  setSort: (sort: SortState) => void
  setPagination: (pagination: Partial<PaginationState>) => void
}

export const useClustersStore = create<ClustersStore>()(
  devtools(
    (set) => ({
      clusters: [],
      selectedCluster: null,
      loading: false,
      error: null,
      filter: {},
      sort: { field: 'name', direction: 'asc' },
      pagination: { page: 1, limit: 20, total: 0, totalPages: 0 },

      setClusters: (clusters) => set({ clusters }),
      addCluster: (cluster) =>
        set((state) => ({ clusters: [...state.clusters, cluster] })),
      updateCluster: (id, updates) =>
        set((state) => ({
          clusters: state.clusters.map((c) =>
            c.id === id ? { ...c, ...updates } : c
          ),
          selectedCluster:
            state.selectedCluster?.id === id
              ? { ...state.selectedCluster, ...updates }
              : state.selectedCluster,
        })),
      removeCluster: (id) =>
        set((state) => ({
          clusters: state.clusters.filter((c) => c.id !== id),
          selectedCluster:
            state.selectedCluster?.id === id ? null : state.selectedCluster,
        })),
      selectCluster: (cluster) => set({ selectedCluster: cluster }),
      setLoading: (loading) => set({ loading }),
      setError: (error) => set({ error }),
      setFilter: (filter) =>
        set((state) => ({ filter: { ...state.filter, ...filter } })),
      setSort: (sort) => set({ sort }),
      setPagination: (pagination) =>
        set((state) => ({ pagination: { ...state.pagination, ...pagination } })),
    }),
    { name: 'clusters-store' }
  )
)

// ============================================================================
// Applications Store
// ============================================================================

interface ApplicationsStore {
  applications: Application[]
  selectedApp: Application | null
  loading: boolean
  error: string | null
  filter: FilterState
  sort: SortState
  pagination: PaginationState

  setApplications: (applications: Application[]) => void
  addApplication: (app: Application) => void
  updateApplication: (id: string, updates: Partial<Application>) => void
  removeApplication: (id: string) => void
  selectApplication: (app: Application | null) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  setFilter: (filter: Partial<FilterState>) => void
  setSort: (sort: SortState) => void
  setPagination: (pagination: Partial<PaginationState>) => void
}

export const useApplicationsStore = create<ApplicationsStore>()(
  devtools(
    (set) => ({
      applications: [],
      selectedApp: null,
      loading: false,
      error: null,
      filter: {},
      sort: { field: 'name', direction: 'asc' },
      pagination: { page: 1, limit: 20, total: 0, totalPages: 0 },

      setApplications: (applications) => set({ applications }),
      addApplication: (app) =>
        set((state) => ({ applications: [...state.applications, app] })),
      updateApplication: (id, updates) =>
        set((state) => ({
          applications: state.applications.map((a) =>
            a.id === id ? { ...a, ...updates } : a
          ),
          selectedApp:
            state.selectedApp?.id === id
              ? { ...state.selectedApp, ...updates }
              : state.selectedApp,
        })),
      removeApplication: (id) =>
        set((state) => ({
          applications: state.applications.filter((a) => a.id !== id),
          selectedApp:
            state.selectedApp?.id === id ? null : state.selectedApp,
        })),
      selectApplication: (app) => set({ selectedApp: app }),
      setLoading: (loading) => set({ loading }),
      setError: (error) => set({ error }),
      setFilter: (filter) =>
        set((state) => ({ filter: { ...state.filter, ...filter } })),
      setSort: (sort) => set({ sort }),
      setPagination: (pagination) =>
        set((state) => ({ pagination: { ...state.pagination, ...pagination } })),
    }),
    { name: 'applications-store' }
  )
)

// ============================================================================
// Pipelines Store
// ============================================================================

interface PipelinesStore {
  pipelines: Pipeline[]
  selectedPipeline: Pipeline | null
  loading: boolean
  error: string | null
  filter: FilterState
  sort: SortState
  pagination: PaginationState

  setPipelines: (pipelines: Pipeline[]) => void
  addPipeline: (pipeline: Pipeline) => void
  updatePipeline: (id: string, updates: Partial<Pipeline>) => void
  removePipeline: (id: string) => void
  selectPipeline: (pipeline: Pipeline | null) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  setFilter: (filter: Partial<FilterState>) => void
  setSort: (sort: SortState) => void
  setPagination: (pagination: Partial<PaginationState>) => void
}

export const usePipelinesStore = create<PipelinesStore>()(
  devtools(
    (set) => ({
      pipelines: [],
      selectedPipeline: null,
      loading: false,
      error: null,
      filter: {},
      sort: { field: 'name', direction: 'asc' },
      pagination: { page: 1, limit: 20, total: 0, totalPages: 0 },

      setPipelines: (pipelines) => set({ pipelines }),
      addPipeline: (pipeline) =>
        set((state) => ({ pipelines: [...state.pipelines, pipeline] })),
      updatePipeline: (id, updates) =>
        set((state) => ({
          pipelines: state.pipelines.map((p) =>
            p.id === id ? { ...p, ...updates } : p
          ),
          selectedPipeline:
            state.selectedPipeline?.id === id
              ? { ...state.selectedPipeline, ...updates }
              : state.selectedPipeline,
        })),
      removePipeline: (id) =>
        set((state) => ({
          pipelines: state.pipelines.filter((p) => p.id !== id),
          selectedPipeline:
            state.selectedPipeline?.id === id ? null : state.selectedPipeline,
        })),
      selectPipeline: (pipeline) => set({ selectedPipeline: pipeline }),
      setLoading: (loading) => set({ loading }),
      setError: (error) => set({ error }),
      setFilter: (filter) =>
        set((state) => ({ filter: { ...state.filter, ...filter } })),
      setSort: (sort) => set({ sort }),
      setPagination: (pagination) =>
        set((state) => ({ pagination: { ...state.pagination, ...pagination } })),
    }),
    { name: 'pipelines-store' }
  )
)

// ============================================================================
// Alerts Store
// ============================================================================

interface AlertsStore {
  alerts: Alert[]
  unreadCount: number
  loading: boolean
  error: string | null
  filter: FilterState

  setAlerts: (alerts: Alert[]) => void
  addAlert: (alert: Alert) => void
  updateAlert: (id: string, updates: Partial<Alert>) => void
  removeAlert: (id: string) => void
  acknowledgeAlert: (id: string, userId: string) => void
  resolveAlert: (id: string, userId: string) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  setFilter: (filter: Partial<FilterState>) => void
  markAllAsRead: () => void
}

export const useAlertsStore = create<AlertsStore>()(
  devtools(
    (set, get) => ({
      alerts: [],
      unreadCount: 0,
      loading: false,
      error: null,
      filter: {},

      setAlerts: (alerts) => {
        const unreadCount = alerts.filter(
          (a) => a.status === 'firing'
        ).length
        set({ alerts, unreadCount })
      },
      addAlert: (alert) =>
        set((state) => ({
          alerts: [alert, ...state.alerts],
          unreadCount:
            alert.status === 'firing'
              ? state.unreadCount + 1
              : state.unreadCount,
        })),
      updateAlert: (id, updates) =>
        set((state) => ({
          alerts: state.alerts.map((a) =>
            a.id === id ? { ...a, ...updates } : a
          ),
        })),
      removeAlert: (id) =>
        set((state) => {
          const alert = state.alerts.find((a) => a.id === id)
          return {
            alerts: state.alerts.filter((a) => a.id !== id),
            unreadCount:
              alert?.status === 'firing'
                ? state.unreadCount - 1
                : state.unreadCount,
          }
        }),
      acknowledgeAlert: (id, userId) =>
        set((state) => ({
          alerts: state.alerts.map((a) =>
            a.id === id
              ? {
                  ...a,
                  status: 'acknowledged' as const,
                  acknowledgedAt: new Date().toISOString(),
                  acknowledgedBy: userId,
                }
              : a
          ),
          unreadCount: Math.max(0, state.unreadCount - 1),
        })),
      resolveAlert: (id, userId) =>
        set((state) => ({
          alerts: state.alerts.map((a) =>
            a.id === id
              ? {
                  ...a,
                  status: 'resolved' as const,
                  resolvedAt: new Date().toISOString(),
                  resolvedBy: userId,
                }
              : a
          ),
        })),
      setLoading: (loading) => set({ loading }),
      setError: (error) => set({ error }),
      setFilter: (filter) =>
        set((state) => ({ filter: { ...state.filter, ...filter } })),
      markAllAsRead: () => set({ unreadCount: 0 }),
    }),
    { name: 'alerts-store' }
  )
)

// ============================================================================
// Cost Store
// ============================================================================

interface CostStore {
  summary: CostSummary | null
  loading: boolean
  error: string | null
  dateRange: {
    start: string
    end: string
  }

  setSummary: (summary: CostSummary) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  setDateRange: (start: string, end: string) => void
}

export const useCostStore = create<CostStore>()(
  devtools(
    (set) => ({
      summary: null,
      loading: false,
      error: null,
      dateRange: {
        start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString(),
      },

      setSummary: (summary) => set({ summary }),
      setLoading: (loading) => set({ loading }),
      setError: (error) => set({ error }),
      setDateRange: (start, end) => set({ dateRange: { start, end } }),
    }),
    { name: 'cost-store' }
  )
)

// ============================================================================
// UI Store
// ============================================================================

interface UIStore {
  sidebarCollapsed: boolean
  theme: 'dark' | 'light'
  commandPaletteOpen: boolean
  notificationPanelOpen: boolean
  currentModal: string | null
  modalData: unknown

  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setTheme: (theme: 'dark' | 'light') => void
  toggleCommandPalette: () => void
  setCommandPaletteOpen: (open: boolean) => void
  toggleNotificationPanel: () => void
  setNotificationPanelOpen: (open: boolean) => void
  openModal: (modal: string, data?: unknown) => void
  closeModal: () => void
}

export const useUIStore = create<UIStore>()(
  persist(
    devtools(
      (set) => ({
        sidebarCollapsed: false,
        theme: 'dark',
        commandPaletteOpen: false,
        notificationPanelOpen: false,
        currentModal: null,
        modalData: null,

        toggleSidebar: () =>
          set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
        setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),
        setTheme: (theme) => set({ theme }),
        toggleCommandPalette: () =>
          set((state) => ({ commandPaletteOpen: !state.commandPaletteOpen })),
        setCommandPaletteOpen: (open) => set({ commandPaletteOpen: open }),
        toggleNotificationPanel: () =>
          set((state) => ({
            notificationPanelOpen: !state.notificationPanelOpen,
          })),
        setNotificationPanelOpen: (open) => set({ notificationPanelOpen: open }),
        openModal: (modal, data) => set({ currentModal: modal, modalData: data }),
        closeModal: () => set({ currentModal: null, modalData: null }),
      }),
      { name: 'ui-store' }
    ),
    {
      name: 'krustron-ui',
      partialize: (state) => ({
        sidebarCollapsed: state.sidebarCollapsed,
        theme: state.theme,
      }),
    }
  )
)

// ============================================================================
// Export all stores
// ============================================================================

export {
  type AuthStore,
  type ClustersStore,
  type ApplicationsStore,
  type PipelinesStore,
  type AlertsStore,
  type CostStore,
  type UIStore,
}
