"use client"

import * as React from "react"
import { Toast } from "./toast"

export type ToastPosition = "top-left" | "top-center" | "top-right" | "bottom-left" | "bottom-center" | "bottom-right"

export interface ToastData {
  id: string
  title?: string
  description?: string
  variant?: "default" | "success" | "destructive" | "warning" | "info"
  action?: {
    label: string
    onClick: () => void
  }
  duration?: number
  custom?: React.ReactNode
  position?: ToastPosition
}

interface ToastContextValue {
  toasts: ToastData[]
  addToast: (toast: Omit<ToastData, "id">) => string
  removeToast: (id: string) => void
  updateToast: (id: string, toast: Partial<ToastData>) => void
}

const ToastContext = React.createContext<ToastContextValue | undefined>(undefined)

export function useToast() {
  const context = React.useContext(ToastContext)
  if (!context) {
    throw new Error("useToast must be used within a ToastProvider")
  }
  return context
}

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = React.useState<ToastData[]>([])

  const addToast = React.useCallback((toast: Omit<ToastData, "id">) => {
    const id = Math.random().toString(36).substr(2, 9)
    const newToast = { ...toast, id }
    setToasts((prev) => [...prev, newToast])

    // Auto remove after duration
    const duration = toast.duration ?? 4000
    if (duration > 0) {
      setTimeout(() => {
        removeToast(id)
      }, duration)
    }

    return id
  }, [])

  const removeToast = React.useCallback((id: string) => {
    // Find the toast to get its position
    const toast = toasts.find(t => t.id === id)
    const position = toast?.position || "bottom-right"
    
    // First mark the toast for removal (triggers exit animation)
    const toastElement = document.getElementById(`toast-${id}`)
    if (toastElement) {
      toastElement.classList.remove("animate-slide-up", "animate-slide-down")
      toastElement.classList.add(position.includes("top") ? "animate-out-up" : "animate-out-down")
      
      // Remove from state after animation completes
      setTimeout(() => {
        setToasts((prev) => prev.filter((toast) => toast.id !== id))
      }, 200)
    } else {
      setToasts((prev) => prev.filter((toast) => toast.id !== id))
    }
  }, [toasts])

  const updateToast = React.useCallback((id: string, updatedToast: Partial<ToastData>) => {
    setToasts((prev) =>
      prev.map((toast) =>
        toast.id === id ? { ...toast, ...updatedToast } : toast
      )
    )
    
    // If duration is updated and greater than 0, set new timeout
    if (updatedToast.duration && updatedToast.duration > 0) {
      setTimeout(() => {
        removeToast(id)
      }, updatedToast.duration)
    }
  }, [removeToast])

  // Group toasts by position
  const toastsByPosition = toasts.reduce((acc, toast) => {
    const position = toast.position || "bottom-right"
    if (!acc[position]) acc[position] = []
    acc[position].push(toast)
    return acc
  }, {} as Record<ToastPosition, ToastData[]>)

  const positionClasses: Record<ToastPosition, string> = {
    "top-left": "fixed top-0 left-0",
    "top-center": "fixed top-0 left-1/2 -translate-x-1/2",
    "top-right": "fixed top-0 right-0",
    "bottom-left": "fixed bottom-0 left-0",
    "bottom-center": "fixed bottom-0 left-1/2 -translate-x-1/2",
    "bottom-right": "fixed bottom-0 right-0",
  }

  return (
    <ToastContext.Provider value={{ toasts, addToast, removeToast, updateToast }}>
      {children}
      {Object.entries(toastsByPosition).map(([position, positionToasts]) => (
        <div
          key={position}
          className={`${positionClasses[position as ToastPosition]} z-[100] flex max-h-screen w-full flex-col p-4 sm:max-w-[420px] ${
            position.includes("top") ? "flex-col" : "flex-col-reverse"
          }`}
        >
          <div className="flex flex-col gap-2 transition-all duration-300" data-toast-container>
            {positionToasts.map((toast, index) => 
              toast.custom ? (
                <div
                  key={toast.id}
                  id={`toast-${toast.id}`}
                  className={`pointer-events-auto transition-all duration-300 ${
                    position.includes("top") ? "animate-slide-down" : "animate-slide-up"
                  }`}
                  style={{
                    transitionDelay: `${index * 50}ms`
                  }}
                >
                  {toast.custom}
                </div>
              ) : (
                <Toast
                  key={toast.id}
                  id={`toast-${toast.id}`}
                  variant={toast.variant}
                  title={toast.title}
                  description={toast.description}
                  onClose={() => removeToast(toast.id)}
                  action={
                    toast.action && (
                      <button
                        onClick={toast.action.onClick}
                        className="inline-flex items-center justify-center text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 border border-current/20 bg-background/10 backdrop-blur-sm hover:bg-background/20 h-8 px-3 rounded-md shrink-0 ml-auto"
                      >
                        {toast.action.label}
                      </button>
                    )
                  }
                  className={`transition-all duration-300 ${
                    position.includes("top") ? "animate-slide-down" : "animate-slide-up"
                  }`}
                  style={{
                    transitionDelay: `${index * 50}ms`
                  }}
                />
              )
            )}
          </div>
        </div>
      ))}
    </ToastContext.Provider>
  )
}