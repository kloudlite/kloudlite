"use client"

import { useToast as useToastContext } from "./toast-provider"

// Store the toast context globally to allow toast calls from anywhere
let globalAddToast: ((toast: any) => string) | null = null
let globalUpdateToast: ((id: string, toast: any) => void) | null = null

export function useToastSetup() {
  const { addToast, updateToast } = useToastContext()
  globalAddToast = addToast
  globalUpdateToast = updateToast
}

// Main toast function with same interface as sonner
function createToast(message: string, options?: any) {
  if (!globalAddToast) {
    console.warn("Toast system not initialized. Make sure ToastProvider is in your app.")
    return ""
  }

  return globalAddToast({
    title: message,
    description: options?.description,
    variant: "default",
    action: options?.action,
    duration: options?.duration,
    position: options?.position,
  })
}

// Success toast
createToast.success = (message: string, options?: any) => {
  if (!globalAddToast) return ""
  return globalAddToast({
    title: message,
    description: options?.description,
    variant: "success",
    action: options?.action,
    duration: options?.duration,
    position: options?.position,
  })
}

// Error toast
createToast.error = (message: string, options?: any) => {
  if (!globalAddToast) return ""
  return globalAddToast({
    title: message,
    description: options?.description,
    variant: "destructive",
    action: options?.action,
    duration: options?.duration,
    position: options?.position,
  })
}

// Warning toast
createToast.warning = (message: string, options?: any) => {
  if (!globalAddToast) return ""
  return globalAddToast({
    title: message,
    description: options?.description,
    variant: "warning",
    action: options?.action,
    duration: options?.duration,
    position: options?.position,
  })
}

// Info toast
createToast.info = (message: string, options?: any) => {
  if (!globalAddToast) return ""
  return globalAddToast({
    title: message,
    description: options?.description,
    variant: "info",
    action: options?.action,
    duration: options?.duration,
    position: options?.position,
  })
}

// Promise toast
createToast.promise = <T,>(
  promise: Promise<T>,
  options: {
    loading: string
    success: string | ((data: T) => string)
    error: string | ((error: any) => string)
  }
) => {
  if (!globalAddToast || !globalUpdateToast) return promise

  const id = globalAddToast({
    title: options.loading,
    variant: "default",
    duration: 100000, // Don't auto-dismiss while loading
  })

  promise
    .then((data) => {
      const successMessage = typeof options.success === "function" 
        ? options.success(data) 
        : options.success
      
      globalUpdateToast!(id, {
        title: successMessage,
        variant: "success",
        duration: 4000,
      })
    })
    .catch((error) => {
      const errorMessage = typeof options.error === "function"
        ? options.error(error)
        : options.error
      
      globalUpdateToast!(id, {
        title: errorMessage,
        variant: "destructive",
        duration: 4000,
      })
    })

  return promise
}

// Update toast
createToast.update = (id: string, options: any) => {
  if (!globalUpdateToast) return
  
  globalUpdateToast(id, {
    title: options.title,
    description: options.description,
    variant: options.variant,
    action: options.action,
    duration: options.duration,
  })
}

// Custom toast
createToast.custom = (content: React.ReactNode) => {
  if (!globalAddToast) return ""
  
  return globalAddToast({
    custom: content,
    duration: 4000,
  })
}

export const toast = createToast