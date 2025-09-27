'use client'

import { useEffect } from 'react'

import { Home, RotateCw } from "lucide-react"
import Link from "next/link"

import { ThemeToggle } from "@/components/theme-toggle"
import { Button } from "@/components/ui/button"

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    console.error(error)
  }, [error])

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <div className="absolute top-4 right-4">
        <ThemeToggle />
      </div>
      
      <div className="flex flex-1 items-center justify-center p-6">
        <div className="w-full max-w-md space-y-8 text-center">
          <div className="space-y-2">
            <h1 className="text-6xl font-light tracking-tighter text-foreground/80">
              500
            </h1>
            <h2 className="text-xl font-medium">Something went wrong</h2>
            <p className="text-sm text-muted-foreground">
              An unexpected error occurred. Please try again.
            </p>
          </div>
          
          <div className="flex flex-col gap-2 sm:flex-row sm:justify-center">
            <Button onClick={reset} size="sm">
              <RotateCw className="mr-2 h-3.5 w-3.5" />
              Try again
            </Button>
            <Button variant="outline" size="sm" asChild>
              <Link href="/">
                <Home className="mr-2 h-3.5 w-3.5" />
                Home
              </Link>
            </Button>
          </div>
          
          {error.digest && (
            <p className="mt-8 text-xs text-muted-foreground">
              Error ID: <code className="font-mono">{error.digest}</code>
            </p>
          )}
        </div>
      </div>
    </div>
  )
}