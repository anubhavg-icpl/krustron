// Krustron Dashboard - Notification Panel Component
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  X,
  Bell,
  AlertTriangle,
  AlertCircle,
  Info,
  CheckCircle,
  XCircle,
  Server,
  Box,
  GitBranch,
  Shield,
  DollarSign,
  Settings,
  Check,
  Trash2,
} from 'lucide-react'
import { clsx } from 'clsx'
import { useAlertsStore, useUIStore } from '@/store/useStore'
import { Alert, AlertSeverity, AlertType } from '@/types'

// Severity icon mapping
function getSeverityIcon(severity: AlertSeverity) {
  switch (severity) {
    case 'critical':
      return XCircle
    case 'high':
      return AlertTriangle
    case 'medium':
      return AlertCircle
    case 'low':
      return Info
    case 'info':
      return Info
    default:
      return Bell
  }
}

// Severity styles
function getSeverityStyles(severity: AlertSeverity) {
  switch (severity) {
    case 'critical':
      return 'bg-status-error/20 text-status-error border-status-error/30'
    case 'high':
      return 'bg-accent-500/20 text-accent-400 border-accent-500/30'
    case 'medium':
      return 'bg-status-warning/20 text-status-warning border-status-warning/30'
    case 'low':
      return 'bg-status-info/20 text-status-info border-status-info/30'
    case 'info':
      return 'bg-primary-500/20 text-primary-400 border-primary-500/30'
    default:
      return 'bg-gray-500/20 text-gray-400 border-gray-500/30'
  }
}

// Type icon mapping
function getTypeIcon(type: AlertType) {
  switch (type) {
    case 'cluster':
      return Server
    case 'application':
      return Box
    case 'pipeline':
      return GitBranch
    case 'security':
      return Shield
    case 'cost':
      return DollarSign
    case 'system':
      return Settings
    default:
      return Bell
  }
}

// Format relative time
function formatRelativeTime(date: string): string {
  const now = new Date()
  const then = new Date(date)
  const diffMs = now.getTime() - then.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  return `${diffDays}d ago`
}

// Single notification item
function NotificationItem({ alert, onAcknowledge, onResolve, onRemove }: {
  alert: Alert
  onAcknowledge: (id: string) => void
  onResolve: (id: string) => void
  onRemove: (id: string) => void
}) {
  const SeverityIcon = getSeverityIcon(alert.severity)
  const TypeIcon = getTypeIcon(alert.type)

  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: 20 }}
      className={clsx(
        'p-4 border-l-4 rounded-r-lg mb-3 transition-all',
        alert.status === 'firing' ? 'bg-glass-light' : 'bg-glass-light/50',
        alert.severity === 'critical' && 'border-status-error',
        alert.severity === 'high' && 'border-accent-500',
        alert.severity === 'medium' && 'border-status-warning',
        alert.severity === 'low' && 'border-status-info',
        alert.severity === 'info' && 'border-primary-500'
      )}
    >
      <div className="flex items-start gap-3">
        {/* Icon */}
        <div className={clsx(
          'p-2 rounded-lg',
          getSeverityStyles(alert.severity)
        )}>
          <SeverityIcon className="w-4 h-4" />
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className={clsx(
              'text-xs px-2 py-0.5 rounded-full border',
              getSeverityStyles(alert.severity)
            )}>
              {alert.severity.toUpperCase()}
            </span>
            <span className="text-xs text-gray-500 flex items-center gap-1">
              <TypeIcon className="w-3 h-3" />
              {alert.type}
            </span>
          </div>

          <h4 className="text-sm font-medium text-white truncate">
            {alert.title}
          </h4>
          <p className="text-xs text-gray-400 mt-1 line-clamp-2">
            {alert.message}
          </p>

          <div className="flex items-center justify-between mt-2">
            <span className="text-xs text-gray-500">
              {formatRelativeTime(alert.startsAt)}
            </span>

            {/* Status badge */}
            <span className={clsx(
              'text-xs px-2 py-0.5 rounded-full',
              alert.status === 'firing' && 'bg-status-error/20 text-status-error',
              alert.status === 'acknowledged' && 'bg-status-warning/20 text-status-warning',
              alert.status === 'resolved' && 'bg-status-healthy/20 text-status-healthy',
              alert.status === 'silenced' && 'bg-gray-500/20 text-gray-400'
            )}>
              {alert.status}
            </span>
          </div>

          {/* Actions */}
          {alert.status === 'firing' && (
            <div className="flex items-center gap-2 mt-3">
              <button
                onClick={() => onAcknowledge(alert.id)}
                className="text-xs px-2 py-1 rounded bg-status-warning/20 text-status-warning hover:bg-status-warning/30 transition-colors"
              >
                Acknowledge
              </button>
              <button
                onClick={() => onResolve(alert.id)}
                className="text-xs px-2 py-1 rounded bg-status-healthy/20 text-status-healthy hover:bg-status-healthy/30 transition-colors"
              >
                Resolve
              </button>
            </div>
          )}
        </div>

        {/* Remove button */}
        <button
          onClick={() => onRemove(alert.id)}
          className="p-1 rounded hover:bg-glass-medium transition-colors"
        >
          <X className="w-4 h-4 text-gray-500" />
        </button>
      </div>
    </motion.div>
  )
}

export default function NotificationPanel() {
  const { notificationPanelOpen, setNotificationPanelOpen } = useUIStore()
  const { alerts, acknowledgeAlert, resolveAlert, removeAlert, markAllAsRead } = useAlertsStore()
  const [filter, setFilter] = useState<'all' | AlertSeverity>('all')

  const filteredAlerts = filter === 'all'
    ? alerts
    : alerts.filter(a => a.severity === filter)

  const handleAcknowledge = (id: string) => {
    acknowledgeAlert(id, 'current-user')
  }

  const handleResolve = (id: string) => {
    resolveAlert(id, 'current-user')
  }

  const severityCounts = {
    critical: alerts.filter(a => a.severity === 'critical' && a.status === 'firing').length,
    high: alerts.filter(a => a.severity === 'high' && a.status === 'firing').length,
    medium: alerts.filter(a => a.severity === 'medium' && a.status === 'firing').length,
    low: alerts.filter(a => a.severity === 'low' && a.status === 'firing').length,
    info: alerts.filter(a => a.severity === 'info' && a.status === 'firing').length,
  }

  return (
    <AnimatePresence>
      {notificationPanelOpen && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/40 backdrop-blur-sm z-40"
            onClick={() => setNotificationPanelOpen(false)}
          />

          {/* Panel */}
          <motion.div
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', damping: 25, stiffness: 300 }}
            className="fixed right-0 top-0 bottom-0 w-full sm:w-96 bg-surface-100 border-l border-glass-border z-50 flex flex-col"
          >
            {/* Header */}
            <div className="h-16 flex items-center justify-between px-4 border-b border-glass-border">
              <div className="flex items-center gap-3">
                <Bell className="w-5 h-5 text-primary-400" />
                <h2 className="text-lg font-semibold text-white">Notifications</h2>
                {alerts.filter(a => a.status === 'firing').length > 0 && (
                  <span className="px-2 py-0.5 bg-status-error rounded-full text-xs text-white">
                    {alerts.filter(a => a.status === 'firing').length}
                  </span>
                )}
              </div>
              <button
                onClick={() => setNotificationPanelOpen(false)}
                className="p-2 rounded-lg hover:bg-glass-light transition-colors"
              >
                <X className="w-5 h-5 text-gray-400" />
              </button>
            </div>

            {/* Severity Filters */}
            <div className="px-4 py-3 border-b border-glass-border">
              <div className="flex flex-wrap gap-2">
                <button
                  onClick={() => setFilter('all')}
                  className={clsx(
                    'px-3 py-1.5 rounded-lg text-xs font-medium transition-colors',
                    filter === 'all'
                      ? 'bg-primary-500/20 text-primary-400'
                      : 'text-gray-400 hover:bg-glass-light'
                  )}
                >
                  All ({alerts.length})
                </button>
                {severityCounts.critical > 0 && (
                  <button
                    onClick={() => setFilter('critical')}
                    className={clsx(
                      'px-3 py-1.5 rounded-lg text-xs font-medium transition-colors flex items-center gap-1',
                      filter === 'critical'
                        ? 'bg-status-error/20 text-status-error'
                        : 'text-gray-400 hover:bg-glass-light'
                    )}
                  >
                    <XCircle className="w-3 h-3" />
                    Critical ({severityCounts.critical})
                  </button>
                )}
                {severityCounts.high > 0 && (
                  <button
                    onClick={() => setFilter('high')}
                    className={clsx(
                      'px-3 py-1.5 rounded-lg text-xs font-medium transition-colors flex items-center gap-1',
                      filter === 'high'
                        ? 'bg-accent-500/20 text-accent-400'
                        : 'text-gray-400 hover:bg-glass-light'
                    )}
                  >
                    <AlertTriangle className="w-3 h-3" />
                    High ({severityCounts.high})
                  </button>
                )}
                {severityCounts.medium > 0 && (
                  <button
                    onClick={() => setFilter('medium')}
                    className={clsx(
                      'px-3 py-1.5 rounded-lg text-xs font-medium transition-colors flex items-center gap-1',
                      filter === 'medium'
                        ? 'bg-status-warning/20 text-status-warning'
                        : 'text-gray-400 hover:bg-glass-light'
                    )}
                  >
                    <AlertCircle className="w-3 h-3" />
                    Medium ({severityCounts.medium})
                  </button>
                )}
                {(severityCounts.low > 0 || severityCounts.info > 0) && (
                  <button
                    onClick={() => setFilter('info')}
                    className={clsx(
                      'px-3 py-1.5 rounded-lg text-xs font-medium transition-colors flex items-center gap-1',
                      filter === 'info' || filter === 'low'
                        ? 'bg-status-info/20 text-status-info'
                        : 'text-gray-400 hover:bg-glass-light'
                    )}
                  >
                    <Info className="w-3 h-3" />
                    Info ({severityCounts.low + severityCounts.info})
                  </button>
                )}
              </div>
            </div>

            {/* Actions Bar */}
            {alerts.length > 0 && (
              <div className="px-4 py-2 border-b border-glass-border flex items-center justify-between">
                <button
                  onClick={markAllAsRead}
                  className="text-xs text-primary-400 hover:text-primary-300 flex items-center gap-1"
                >
                  <Check className="w-3 h-3" />
                  Mark all as read
                </button>
              </div>
            )}

            {/* Notifications List */}
            <div className="flex-1 overflow-y-auto p-4">
              {filteredAlerts.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-gray-400">
                  <Bell className="w-12 h-12 mb-3 opacity-50" />
                  <p className="text-sm font-medium">No notifications</p>
                  <p className="text-xs mt-1">You're all caught up!</p>
                </div>
              ) : (
                <AnimatePresence>
                  {filteredAlerts.map((alert) => (
                    <NotificationItem
                      key={alert.id}
                      alert={alert}
                      onAcknowledge={handleAcknowledge}
                      onResolve={handleResolve}
                      onRemove={removeAlert}
                    />
                  ))}
                </AnimatePresence>
              )}
            </div>

            {/* Footer */}
            <div className="px-4 py-3 border-t border-glass-border">
              <p className="text-xs text-gray-500 text-center">
                Notifications are synced in real-time
              </p>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}
