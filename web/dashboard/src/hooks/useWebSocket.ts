// Krustron Dashboard - WebSocket Hook
// Author: Anubhav Gain <anubhavg@infopercept.com>
// Real-time communication with the Krustron backend

import { useEffect, useRef, useCallback, useState } from 'react'
import { WebSocketMessage, MessageType } from '@/types'
import { useAuthStore } from '@/store/useStore'

// ============================================================================
// Types
// ============================================================================

export interface WebSocketOptions {
  url?: string
  reconnect?: boolean
  reconnectAttempts?: number
  heartbeatInterval?: number
  onOpen?: () => void
  onClose?: () => void
  onError?: (error: Event) => void
  onMessage?: (message: WebSocketMessage) => void
  onAuthFailure?: () => void
}

export interface WebSocketState {
  isConnected: boolean
  isConnecting: boolean
  isAuthenticated: boolean
  lastMessage: WebSocketMessage | null
  error: Event | null
  reconnectCount: number
}

export interface UseWebSocketReturn extends WebSocketState {
  send: (message: Partial<WebSocketMessage>) => void
  subscribe: (channel: string, callback: (message: WebSocketMessage) => void) => () => void
  unsubscribe: (channel: string) => void
  disconnect: () => void
  reconnect: () => void
}

// ============================================================================
// Constants
// ============================================================================

const DEFAULT_OPTIONS: Required<Omit<WebSocketOptions, 'url' | 'onOpen' | 'onClose' | 'onError' | 'onMessage' | 'onAuthFailure'>> = {
  reconnect: true,
  reconnectAttempts: 10,
  heartbeatInterval: 30000,
}

// Exponential backoff constants
const INITIAL_RECONNECT_DELAY = 1000
const MAX_RECONNECT_DELAY = 30000

// ============================================================================
// Exponential Backoff Calculator
// ============================================================================

function calculateReconnectDelay(attempt: number): number {
  // Exponential backoff with jitter
  const exponentialDelay = INITIAL_RECONNECT_DELAY * Math.pow(2, attempt)
  const jitter = Math.random() * 1000
  return Math.min(MAX_RECONNECT_DELAY, exponentialDelay + jitter)
}

// ============================================================================
// Hook Implementation
// ============================================================================

export function useWebSocket(options: WebSocketOptions = {}): UseWebSocketReturn {
  const {
    url = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`,
    reconnect = DEFAULT_OPTIONS.reconnect,
    reconnectAttempts = DEFAULT_OPTIONS.reconnectAttempts,
    heartbeatInterval = DEFAULT_OPTIONS.heartbeatInterval,
    onOpen,
    onClose,
    onError,
    onMessage,
    onAuthFailure,
  } = options

  // State
  const [state, setState] = useState<WebSocketState>({
    isConnected: false,
    isConnecting: false,
    isAuthenticated: false,
    lastMessage: null,
    error: null,
    reconnectCount: 0,
  })

  // Refs
  const socketRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const heartbeatIntervalRef = useRef<NodeJS.Timeout | null>(null)
  const subscriptionsRef = useRef<Map<string, Set<(message: WebSocketMessage) => void>>>(new Map())
  const reconnectCountRef = useRef(0)

  // Generate unique message ID
  const generateId = useCallback(() => {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
  }, [])

  // Clear timers
  const clearTimers = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    if (heartbeatIntervalRef.current) {
      clearInterval(heartbeatIntervalRef.current)
      heartbeatIntervalRef.current = null
    }
  }, [])

  // Start heartbeat
  const startHeartbeat = useCallback(() => {
    if (heartbeatIntervalRef.current) {
      clearInterval(heartbeatIntervalRef.current)
    }

    heartbeatIntervalRef.current = setInterval(() => {
      if (socketRef.current?.readyState === WebSocket.OPEN) {
        socketRef.current.send(JSON.stringify({
          id: generateId(),
          type: 'ping',
          channel: 'system',
          payload: null,
          timestamp: new Date().toISOString(),
        }))
      }
    }, heartbeatInterval)
  }, [heartbeatInterval, generateId])

  // Send authentication message
  const sendAuth = useCallback((socket: WebSocket) => {
    const { token } = useAuthStore.getState()
    if (!token) {
      console.error('[WebSocket] No token available for authentication')
      onAuthFailure?.()
      return
    }

    socket.send(JSON.stringify({
      id: generateId(),
      type: 'auth',
      channel: 'system',
      payload: { token },
      timestamp: new Date().toISOString(),
    }))
  }, [generateId, onAuthFailure])

  // Handle incoming messages
  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const message: WebSocketMessage = JSON.parse(event.data)

      // Update state
      setState(prev => ({ ...prev, lastMessage: message }))

      // Handle authentication response
      if (message.type === 'auth_success') {
        console.log('[WebSocket] Authentication successful')
        setState(prev => ({ ...prev, isAuthenticated: true }))
        return
      }

      if (message.type === 'auth_failure' || message.type === 'auth_error') {
        console.error('[WebSocket] Authentication failed:', message.payload)
        setState(prev => ({ ...prev, isAuthenticated: false }))
        onAuthFailure?.()
        return
      }

      // Handle pong
      if (message.type === 'pong') {
        return
      }

      // Notify global handler
      onMessage?.(message)

      // Notify channel subscribers
      const channelSubscribers = subscriptionsRef.current.get(message.channel)
      if (channelSubscribers) {
        channelSubscribers.forEach(callback => callback(message))
      }

      // Notify type subscribers (for convenience)
      const typeSubscribers = subscriptionsRef.current.get(message.type)
      if (typeSubscribers) {
        typeSubscribers.forEach(callback => callback(message))
      }

      // Notify wildcard subscribers
      const wildcardSubscribers = subscriptionsRef.current.get('*')
      if (wildcardSubscribers) {
        wildcardSubscribers.forEach(callback => callback(message))
      }
    } catch (error) {
      console.error('[WebSocket] Failed to parse message:', error)
    }
  }, [onMessage, onAuthFailure])

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    setState(prev => ({ ...prev, isConnecting: true, error: null, isAuthenticated: false }))

    try {
      const socket = new WebSocket(url)
      socketRef.current = socket

      socket.onopen = () => {
        console.log('[WebSocket] Connected to', url)
        reconnectCountRef.current = 0
        setState(prev => ({
          ...prev,
          isConnected: true,
          isConnecting: false,
          reconnectCount: 0,
        }))

        // Send authentication immediately after connect
        sendAuth(socket)

        startHeartbeat()
        onOpen?.()

        // Re-subscribe to all channels (after auth message)
        setTimeout(() => {
          subscriptionsRef.current.forEach((_, channel) => {
            if (channel !== '*' && !channel.includes('.')) {
              socket.send(JSON.stringify({
                id: generateId(),
                type: 'subscribe',
                channel,
                payload: null,
                timestamp: new Date().toISOString(),
              }))
            }
          })
        }, 100)
      }

      socket.onclose = (event) => {
        console.log('[WebSocket] Disconnected:', event.code, event.reason)
        clearTimers()
        setState(prev => ({ ...prev, isConnected: false, isConnecting: false, isAuthenticated: false }))
        onClose?.()

        // Attempt reconnection with exponential backoff
        if (reconnect && reconnectCountRef.current < reconnectAttempts) {
          const delay = calculateReconnectDelay(reconnectCountRef.current)
          reconnectCountRef.current++
          setState(prev => ({ ...prev, reconnectCount: reconnectCountRef.current }))
          console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${reconnectCountRef.current}/${reconnectAttempts})`)

          reconnectTimeoutRef.current = setTimeout(() => {
            connect()
          }, delay)
        }
      }

      socket.onerror = (error) => {
        console.error('[WebSocket] Error:', error)
        setState(prev => ({ ...prev, error }))
        onError?.(error)
      }

      socket.onmessage = handleMessage
    } catch (error) {
      console.error('[WebSocket] Failed to connect:', error)
      setState(prev => ({ ...prev, isConnecting: false }))
    }
  }, [url, reconnect, reconnectAttempts, onOpen, onClose, onError, handleMessage, startHeartbeat, clearTimers, generateId, sendAuth])

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    clearTimers()
    if (socketRef.current) {
      socketRef.current.close(1000, 'Client disconnect')
      socketRef.current = null
    }
    setState({
      isConnected: false,
      isConnecting: false,
      isAuthenticated: false,
      lastMessage: null,
      error: null,
      reconnectCount: 0,
    })
  }, [clearTimers])

  // Reconnect
  const reconnectFn = useCallback(() => {
    disconnect()
    reconnectCountRef.current = 0
    connect()
  }, [disconnect, connect])

  // Send message
  const send = useCallback((message: Partial<WebSocketMessage>) => {
    if (socketRef.current?.readyState !== WebSocket.OPEN) {
      console.warn('[WebSocket] Cannot send message: not connected')
      return
    }

    const fullMessage: WebSocketMessage = {
      id: generateId(),
      type: (message.type || 'notification') as MessageType,
      channel: message.channel || 'default',
      payload: message.payload,
      timestamp: new Date().toISOString(),
      metadata: message.metadata,
    }

    socketRef.current.send(JSON.stringify(fullMessage))
  }, [generateId])

  // Subscribe to channel
  const subscribe = useCallback((channel: string, callback: (message: WebSocketMessage) => void) => {
    // Add to local subscriptions
    if (!subscriptionsRef.current.has(channel)) {
      subscriptionsRef.current.set(channel, new Set())
    }
    subscriptionsRef.current.get(channel)!.add(callback)

    // Send subscribe message to server (if it's a channel, not a message type)
    if (socketRef.current?.readyState === WebSocket.OPEN && !channel.includes('.') && channel !== '*') {
      socketRef.current.send(JSON.stringify({
        id: generateId(),
        type: 'subscribe',
        channel,
        payload: null,
        timestamp: new Date().toISOString(),
      }))
    }

    // Return unsubscribe function
    return () => {
      const subscribers = subscriptionsRef.current.get(channel)
      if (subscribers) {
        subscribers.delete(callback)
        if (subscribers.size === 0) {
          subscriptionsRef.current.delete(channel)
          // Unsubscribe from server
          if (socketRef.current?.readyState === WebSocket.OPEN && !channel.includes('.') && channel !== '*') {
            socketRef.current.send(JSON.stringify({
              id: generateId(),
              type: 'unsubscribe',
              channel,
              payload: null,
              timestamp: new Date().toISOString(),
            }))
          }
        }
      }
    }
  }, [generateId])

  // Unsubscribe from channel
  const unsubscribe = useCallback((channel: string) => {
    subscriptionsRef.current.delete(channel)
    if (socketRef.current?.readyState === WebSocket.OPEN && !channel.includes('.') && channel !== '*') {
      socketRef.current.send(JSON.stringify({
        id: generateId(),
        type: 'unsubscribe',
        channel,
        payload: null,
        timestamp: new Date().toISOString(),
      }))
    }
  }, [generateId])

  // Connect on mount
  useEffect(() => {
    connect()
    return () => {
      disconnect()
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return {
    ...state,
    send,
    subscribe,
    unsubscribe,
    disconnect,
    reconnect: reconnectFn,
  }
}

// ============================================================================
// Channel-specific Hooks
// ============================================================================

export function useClusterEvents(clusterId?: string) {
  const { subscribe, isConnected } = useWebSocket()
  const [events, setEvents] = useState<WebSocketMessage[]>([])

  useEffect(() => {
    if (!isConnected) return

    const channel = clusterId ? `cluster:${clusterId}` : 'clusters'
    const unsubscribe = subscribe(channel, (message) => {
      setEvents(prev => [...prev.slice(-99), message])
    })

    return unsubscribe
  }, [clusterId, subscribe, isConnected])

  return events
}

export function useApplicationEvents(appId?: string) {
  const { subscribe, isConnected } = useWebSocket()
  const [events, setEvents] = useState<WebSocketMessage[]>([])

  useEffect(() => {
    if (!isConnected) return

    const channel = appId ? `app:${appId}` : 'applications'
    const unsubscribe = subscribe(channel, (message) => {
      setEvents(prev => [...prev.slice(-99), message])
    })

    return unsubscribe
  }, [appId, subscribe, isConnected])

  return events
}

export function usePipelineEvents(pipelineId?: string) {
  const { subscribe, isConnected } = useWebSocket()
  const [events, setEvents] = useState<WebSocketMessage[]>([])

  useEffect(() => {
    if (!isConnected) return

    const channel = pipelineId ? `pipeline:${pipelineId}` : 'pipelines'
    const unsubscribe = subscribe(channel, (message) => {
      setEvents(prev => [...prev.slice(-99), message])
    })

    return unsubscribe
  }, [pipelineId, subscribe, isConnected])

  return events
}

export function useAlerts() {
  const { subscribe, isConnected } = useWebSocket()
  const [alerts, setAlerts] = useState<WebSocketMessage[]>([])

  useEffect(() => {
    if (!isConnected) return

    const unsubscribe = subscribe('alerts', (message) => {
      if (message.type === 'alert') {
        setAlerts(prev => [...prev.slice(-49), message])
      } else if (message.type === 'alert.resolved') {
        setAlerts(prev => prev.filter(a => a.id !== message.payload))
      }
    })

    return unsubscribe
  }, [subscribe, isConnected])

  return alerts
}

export function usePodLogs(clusterId: string, namespace: string, podName: string) {
  const { subscribe, send, isConnected } = useWebSocket()
  const [logs, setLogs] = useState<string[]>([])

  useEffect(() => {
    if (!isConnected) return

    const channel = `pod:${clusterId}:${namespace}:${podName}`

    // Subscribe to pod logs
    const unsubscribe = subscribe(channel, (message) => {
      if (message.type === 'pod.logs') {
        setLogs(prev => [...prev.slice(-999), message.payload as string])
      }
    })

    // Request logs stream
    send({
      type: 'subscribe' as MessageType,
      channel,
      payload: { follow: true },
    })

    return () => {
      unsubscribe()
      send({
        type: 'unsubscribe' as MessageType,
        channel,
        payload: null,
      })
    }
  }, [clusterId, namespace, podName, subscribe, send, isConnected])

  return logs
}

// ============================================================================
// WebSocket Context (for sharing connection)
// ============================================================================

import React, { createContext, useContext, ReactNode } from 'react'

interface WebSocketContextValue extends UseWebSocketReturn {}

const WebSocketContext = createContext<WebSocketContextValue | null>(null)

export function WebSocketProvider({ children, options }: { children: ReactNode; options?: WebSocketOptions }) {
  const ws = useWebSocket(options)
  return React.createElement(WebSocketContext.Provider, { value: ws }, children)
}

export function useWebSocketContext() {
  const context = useContext(WebSocketContext)
  if (!context) {
    throw new Error('useWebSocketContext must be used within a WebSocketProvider')
  }
  return context
}

export default useWebSocket
