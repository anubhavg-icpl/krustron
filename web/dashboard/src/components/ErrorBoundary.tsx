// Krustron Dashboard - Error Boundary
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { Component, ErrorInfo, ReactNode } from 'react'
import { AlertTriangle, RefreshCw, Home } from 'lucide-react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
    errorInfo: null,
  }

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error, errorInfo: null }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
    this.setState({ errorInfo })
  }

  private handleRetry = () => {
    this.setState({ hasError: false, error: null, errorInfo: null })
  }

  private handleGoHome = () => {
    window.location.href = '/'
  }

  public render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <div className="min-h-screen bg-surface flex items-center justify-center p-4">
          <div className="glass-card p-8 max-w-lg w-full text-center">
            <div className="w-16 h-16 mx-auto rounded-2xl bg-status-error/20 flex items-center justify-center mb-6">
              <AlertTriangle className="w-8 h-8 text-status-error" />
            </div>

            <h1 className="text-2xl font-bold text-white mb-2">Something went wrong</h1>
            <p className="text-gray-400 mb-6">
              An unexpected error occurred. Please try again or go back to the dashboard.
            </p>

            {process.env.NODE_ENV === 'development' && this.state.error && (
              <div className="bg-glass-light rounded-xl p-4 mb-6 text-left">
                <p className="text-sm font-mono text-status-error mb-2">
                  {this.state.error.message}
                </p>
                {this.state.errorInfo && (
                  <pre className="text-xs text-gray-500 overflow-auto max-h-40">
                    {this.state.errorInfo.componentStack}
                  </pre>
                )}
              </div>
            )}

            <div className="flex gap-4 justify-center">
              <button
                onClick={this.handleRetry}
                className="glass-btn flex items-center gap-2"
              >
                <RefreshCw className="w-4 h-4" />
                Try Again
              </button>
              <button
                onClick={this.handleGoHome}
                className="glass-btn-primary flex items-center gap-2"
              >
                <Home className="w-4 h-4" />
                Go to Dashboard
              </button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}

// Smaller error boundary for sections
export function SectionErrorBoundary({ children }: { children: ReactNode }) {
  return (
    <ErrorBoundary
      fallback={
        <div className="glass-card p-6 text-center">
          <AlertTriangle className="w-8 h-8 text-status-error mx-auto mb-3" />
          <p className="text-gray-400">Failed to load this section</p>
          <button
            onClick={() => window.location.reload()}
            className="glass-btn mt-4 text-sm"
          >
            Reload Page
          </button>
        </div>
      }
    >
      {children}
    </ErrorBoundary>
  )
}

export default ErrorBoundary
