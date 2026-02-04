"use client"

import { useState, useEffect } from "react"

/**
 * Hook to detect if component has mounted on the client.
 * Used to prevent hydration mismatches with Radix UI components
 * that generate different IDs on server vs client.
 */
export function useMounted() {
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  return mounted
}
