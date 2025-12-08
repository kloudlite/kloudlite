'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { Terminal, X, Loader2, Download, Trash2, Pause, Play } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { cn } from '@kloudlite/lib'

interface ServiceLogsViewerProps {
  serviceName: string
  namespace: string
  isOpen: boolean
  onClose: () => void
}

export function ServiceLogsViewer({
  serviceName,
  namespace,
  isOpen,
  onClose,
}: ServiceLogsViewerProps) {
  const [logs, setLogs] = useState<string[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const [isConnecting, setIsConnecting] = useState(false)
  const [isPaused, setIsPaused] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const logsContainerRef = useRef<HTMLDivElement>(null)
  const eventSourceRef = useRef<EventSource | null>(null)
  const pausedLogsRef = useRef<string[]>([])

  const scrollToBottom = useCallback(() => {
    if (logsContainerRef.current && !isPaused) {
      logsContainerRef.current.scrollTop = logsContainerRef.current.scrollHeight
    }
  }, [isPaused])

  const connect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    setIsConnecting(true)
    setError(null)

    const url = `/api/v1/namespaces/${encodeURIComponent(namespace)}/services/${encodeURIComponent(serviceName)}/logs?follow=true&tailLines=200`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      setIsConnected(true)
      setIsConnecting(false)
      setError(null)
    }

    eventSource.onmessage = (event) => {
      const line = event.data
      if (isPaused) {
        pausedLogsRef.current.push(line)
      } else {
        setLogs((prev) => [...prev, line])
        // Schedule scroll after state update
        setTimeout(scrollToBottom, 0)
      }
    }

    eventSource.onerror = () => {
      setIsConnected(false)
      setIsConnecting(false)
      if (eventSource.readyState === EventSource.CLOSED) {
        setError('Connection closed')
      } else {
        setError('Connection error')
      }
      eventSource.close()
    }
  }, [namespace, serviceName, isPaused, scrollToBottom])

  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
    setIsConnected(false)
    setIsConnecting(false)
  }, [])

  // Connect when opened
  useEffect(() => {
    if (isOpen && !isConnected && !isConnecting) {
      connect()
    }
    return () => {
      disconnect()
    }
  }, [isOpen, connect, disconnect, isConnected, isConnecting])

  // Handle pause/resume
  useEffect(() => {
    if (!isPaused && pausedLogsRef.current.length > 0) {
      setLogs((prev) => [...prev, ...pausedLogsRef.current])
      pausedLogsRef.current = []
      setTimeout(scrollToBottom, 0)
    }
  }, [isPaused, scrollToBottom])

  const handleClear = () => {
    setLogs([])
    pausedLogsRef.current = []
  }

  const handleDownload = () => {
    const content = logs.join('\n')
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${serviceName}-logs-${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const handleClose = () => {
    disconnect()
    setLogs([])
    setError(null)
    onClose()
  }

  if (!isOpen) return null

  return (
    <div className="bg-background fixed inset-x-0 bottom-0 z-50 flex h-[40vh] min-h-[300px] flex-col border-t shadow-lg">
      {/* Header */}
      <div className="bg-muted/50 flex items-center justify-between border-b px-4 py-2">
        <div className="flex items-center gap-3">
          <Terminal className="h-4 w-4" />
          <span className="text-sm font-medium">
            Logs: {serviceName}
          </span>
          {isConnecting && (
            <span className="text-muted-foreground flex items-center gap-1 text-xs">
              <Loader2 className="h-3 w-3 animate-spin" />
              Connecting...
            </span>
          )}
          {isConnected && (
            <span className="flex items-center gap-1 text-xs text-green-500">
              <span className="h-2 w-2 rounded-full bg-green-500 animate-pulse" />
              Live
            </span>
          )}
          {error && (
            <span className="text-xs text-red-500">{error}</span>
          )}
          {isPaused && (
            <span className="text-muted-foreground text-xs">
              (Paused - {pausedLogsRef.current.length} new lines)
            </span>
          )}
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsPaused(!isPaused)}
            title={isPaused ? 'Resume' : 'Pause'}
          >
            {isPaused ? (
              <Play className="h-4 w-4" />
            ) : (
              <Pause className="h-4 w-4" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleClear}
            title="Clear logs"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleDownload}
            disabled={logs.length === 0}
            title="Download logs"
          >
            <Download className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleClose}
            title="Close"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Logs content */}
      <div
        ref={logsContainerRef}
        className="flex-1 overflow-auto bg-zinc-950 p-4 font-mono text-xs"
      >
        {logs.length === 0 ? (
          <div className="text-muted-foreground flex h-full items-center justify-center">
            {isConnecting ? 'Connecting to log stream...' : 'Waiting for logs...'}
          </div>
        ) : (
          <div className="space-y-0.5">
            {logs.map((line, idx) => (
              <div
                key={idx}
                className={cn(
                  'whitespace-pre-wrap break-all text-zinc-300',
                  line.toLowerCase().includes('error') && 'text-red-400',
                  line.toLowerCase().includes('warn') && 'text-yellow-400',
                  line.toLowerCase().includes('info') && 'text-blue-400',
                )}
              >
                {line}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
