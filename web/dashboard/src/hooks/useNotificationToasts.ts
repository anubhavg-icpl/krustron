// Krustron Dashboard - Notification Toasts Hook
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useEffect, useRef } from 'react'
import toast from 'react-hot-toast'
import { useAlertsStore } from '@/store/useStore'
import { Alert } from '@/types'

// Custom toast styles for different severities
const toastStyles = {
  critical: {
    style: {
      background: 'rgba(239, 68, 68, 0.95)',
      color: '#fff',
      backdropFilter: 'blur(12px)',
      border: '1px solid rgba(255, 255, 255, 0.2)',
      borderRadius: '12px',
    },
    duration: 6000,
  },
  high: {
    style: {
      background: 'rgba(249, 115, 22, 0.95)',
      color: '#fff',
      backdropFilter: 'blur(12px)',
      border: '1px solid rgba(255, 255, 255, 0.2)',
      borderRadius: '12px',
    },
    duration: 5000,
  },
  medium: {
    style: {
      background: 'rgba(234, 179, 8, 0.95)',
      color: '#000',
      backdropFilter: 'blur(12px)',
      border: '1px solid rgba(255, 255, 255, 0.2)',
      borderRadius: '12px',
    },
    duration: 4000,
  },
  low: {
    style: {
      background: 'rgba(59, 130, 246, 0.95)',
      color: '#fff',
      backdropFilter: 'blur(12px)',
      border: '1px solid rgba(255, 255, 255, 0.2)',
      borderRadius: '12px',
    },
    duration: 4000,
  },
  info: {
    style: {
      background: 'rgba(37, 99, 235, 0.95)',
      color: '#fff',
      backdropFilter: 'blur(12px)',
      border: '1px solid rgba(255, 255, 255, 0.2)',
      borderRadius: '12px',
    },
    duration: 3000,
  },
}

const severityIcons = {
  critical: '🚨',
  high: '⚠️',
  medium: '⚡',
  low: 'ℹ️',
  info: '💡',
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
      const style = toastStyles[alert.severity] || toastStyles.info
      const icon = severityIcons[alert.severity] || '🔔'

      toast(
        (t) => (
          <div
            className="flex items-start gap-3 cursor-pointer"
            onClick={() => toast.dismiss(t.id)}
          >
            <span className="text-xl">{icon}</span>
            <div className="flex-1">
              <p className="font-semibold text-sm">{alert.title}</p>
              <p className="text-xs opacity-90 mt-0.5 line-clamp-2">{alert.message}</p>
              <p className="text-xs opacity-75 mt-1">{alert.source} • {alert.type}</p>
            </div>
          </div>
        ),
        style
      )
    })

    // Update ref for next comparison
    previousAlertsRef.current = alerts
  }, [alerts])
}

// Utility functions for showing toasts from anywhere
export const showSuccessToast = (message: string, description?: string) => {
  toast.success(
    (t) => (
      <div className="flex items-start gap-2" onClick={() => toast.dismiss(t.id)}>
        <div>
          <p className="font-semibold text-sm">{message}</p>
          {description && <p className="text-xs opacity-75 mt-0.5">{description}</p>}
        </div>
      </div>
    ),
    {
      duration: 4000,
      style: {
        background: 'rgba(34, 197, 94, 0.95)',
        color: '#fff',
        backdropFilter: 'blur(12px)',
        border: '1px solid rgba(255, 255, 255, 0.2)',
        borderRadius: '12px',
      },
    }
  )
}

export const showErrorToast = (message: string, description?: string) => {
  toast.error(
    (t) => (
      <div className="flex items-start gap-2" onClick={() => toast.dismiss(t.id)}>
        <div>
          <p className="font-semibold text-sm">{message}</p>
          {description && <p className="text-xs opacity-75 mt-0.5">{description}</p>}
        </div>
      </div>
    ),
    {
      duration: 5000,
      style: {
        background: 'rgba(239, 68, 68, 0.95)',
        color: '#fff',
        backdropFilter: 'blur(12px)',
        border: '1px solid rgba(255, 255, 255, 0.2)',
        borderRadius: '12px',
      },
    }
  )
}

export const showInfoToast = (message: string, description?: string) => {
  toast(
    (t) => (
      <div className="flex items-start gap-2" onClick={() => toast.dismiss(t.id)}>
        <div>
          <p className="font-semibold text-sm">{message}</p>
          {description && <p className="text-xs opacity-75 mt-0.5">{description}</p>}
        </div>
      </div>
    ),
    {
      duration: 4000,
      icon: 'ℹ️',
      style: {
        background: 'rgba(59, 130, 246, 0.95)',
        color: '#fff',
        backdropFilter: 'blur(12px)',
        border: '1px solid rgba(255, 255, 255, 0.2)',
        borderRadius: '12px',
      },
    }
  )
}

export const showWarningToast = (message: string, description?: string) => {
  toast(
    (t) => (
      <div className="flex items-start gap-2" onClick={() => toast.dismiss(t.id)}>
        <div>
          <p className="font-semibold text-sm">{message}</p>
          {description && <p className="text-xs opacity-75 mt-0.5">{description}</p>}
        </div>
      </div>
    ),
    {
      duration: 4000,
      icon: '⚠️',
      style: {
        background: 'rgba(234, 179, 8, 0.95)',
        color: '#000',
        backdropFilter: 'blur(12px)',
        border: '1px solid rgba(255, 255, 255, 0.2)',
        borderRadius: '12px',
      },
    }
  )
}
