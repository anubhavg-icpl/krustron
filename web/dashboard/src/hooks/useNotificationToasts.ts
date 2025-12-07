// Krustron Dashboard - Notification Toasts Hook
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useEffect, useRef } from 'react'
import toast from 'react-hot-toast'
import { useAlertsStore } from '@/store/useStore'
import { Alert, AlertSeverity } from '@/types'

// Custom toast styles for different severities
const getToastStyle = (severity: AlertSeverity) => {
  const styles: Record<AlertSeverity, { background: string; color: string }> = {
    critical: { background: 'rgba(239, 68, 68, 0.95)', color: '#fff' },
    high: { background: 'rgba(249, 115, 22, 0.95)', color: '#fff' },
    medium: { background: 'rgba(234, 179, 8, 0.95)', color: '#000' },
    low: { background: 'rgba(59, 130, 246, 0.95)', color: '#fff' },
    info: { background: 'rgba(37, 99, 235, 0.95)', color: '#fff' },
  }
  return {
    ...styles[severity],
    backdropFilter: 'blur(12px)',
    border: '1px solid rgba(255, 255, 255, 0.2)',
    borderRadius: '12px',
  }
}

const getDuration = (severity: AlertSeverity): number => {
  const durations: Record<AlertSeverity, number> = {
    critical: 6000,
    high: 5000,
    medium: 4000,
    low: 4000,
    info: 3000,
  }
  return durations[severity]
}

export function useNotificationToasts() {
  const { alerts } = useAlertsStore()
  const previousAlertsRef = useRef<Alert[]>([])

  useEffect(() => {
    // Find new alerts that weren't in the previous state
    const previousIds = new Set(previousAlertsRef.current.map(a => a.id))
    const newAlerts = alerts.filter(a => !previousIds.has(a.id) && a.status === 'firing')

    // Show toast for each new alert
    newAlerts.forEach(alert => {
      toast(alert.title + (alert.message ? ': ' + alert.message : ''), {
        duration: getDuration(alert.severity),
        style: getToastStyle(alert.severity),
      })
    })

    // Update ref for next comparison
    previousAlertsRef.current = alerts
  }, [alerts])
}

// Utility functions for showing toasts from anywhere
export const showSuccessToast = (message: string, description?: string) => {
  toast.success(description ? `${message}: ${description}` : message, {
    duration: 4000,
    style: {
      background: 'rgba(34, 197, 94, 0.95)',
      color: '#fff',
      backdropFilter: 'blur(12px)',
      border: '1px solid rgba(255, 255, 255, 0.2)',
      borderRadius: '12px',
    },
  })
}

export const showErrorToast = (message: string, description?: string) => {
  toast.error(description ? `${message}: ${description}` : message, {
    duration: 5000,
    style: {
      background: 'rgba(239, 68, 68, 0.95)',
      color: '#fff',
      backdropFilter: 'blur(12px)',
      border: '1px solid rgba(255, 255, 255, 0.2)',
      borderRadius: '12px',
    },
  })
}

export const showInfoToast = (message: string, description?: string) => {
  toast(description ? `${message}: ${description}` : message, {
    duration: 4000,
    style: {
      background: 'rgba(59, 130, 246, 0.95)',
      color: '#fff',
      backdropFilter: 'blur(12px)',
      border: '1px solid rgba(255, 255, 255, 0.2)',
      borderRadius: '12px',
    },
  })
}

export const showWarningToast = (message: string, description?: string) => {
  toast(description ? `${message}: ${description}` : message, {
    duration: 4000,
    style: {
      background: 'rgba(234, 179, 8, 0.95)',
      color: '#000',
      backdropFilter: 'blur(12px)',
      border: '1px solid rgba(255, 255, 255, 0.2)',
      borderRadius: '12px',
    },
  })
}
