'use client'

import { Home, ArrowLeft } from "lucide-react"
import Link from "next/link"

import { ThemeToggle } from "@/components/theme-toggle"
import { Button } from "@/components/ui/button"

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col bg-background">
      <div className="absolute top-4 right-4">
        <ThemeToggle />
      </div>
      
      <div className="flex flex-1 items-center justify-center p-6">
        <div className="w-full max-w-md space-y-8 text-center">
          <div className="space-y-2">
            <h1 className="text-6xl font-light tracking-tighter text-foreground/80">
              404
            </h1>
            <h2 className="text-xl font-medium">Page not found</h2>
            <p className="text-sm text-muted-foreground">
              The page you are looking for doesn't exist or has been moved.
            </p>
          </div>
          
          <div className="flex flex-col gap-2 sm:flex-row sm:justify-center">
            <Button asChild size="sm">
              <Link href="/">
                <Home className="mr-2 h-3.5 w-3.5" />
                Home
              </Link>
            </Button>
            <Button variant="outline" size="sm" onClick={() => window.history.back()}>
              <ArrowLeft className="mr-2 h-3.5 w-3.5" />
              Go back
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}