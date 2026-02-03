'use client'

import { useEffect, useState } from 'react'
import { Moon, Sun, Monitor } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from './dropdown-menu'

export interface ThemeSwitcherProps {
  setThemeCookie?: (theme: 'light' | 'dark' | 'system') => Promise<void>
}

function getCookieValue(name: string): string | null {
  if (typeof document === 'undefined') return null
  const cookies = document.cookie.split(';')
  for (const cookie of cookies) {
    const [key, value] = cookie.trim().split('=')
    if (key === name) return value
  }
  return null
}

export function ThemeSwitcher({ setThemeCookie }: ThemeSwitcherProps) {
  const [mounted, setMounted] = useState(false)
  const [currentTheme, setCurrentTheme] = useState<'light' | 'dark' | 'system'>('system')

  useEffect(() => {
    setMounted(true)
    const theme = getCookieValue('theme') as 'light' | 'dark' | 'system' | null
    setCurrentTheme(theme || 'light')
  }, [])

  // Note: System theme change listener is handled by layout.tsx script globally
  // No need for duplicate listener here to avoid hydration mismatches

  const handleThemeChange = async (newTheme: 'light' | 'dark' | 'system') => {
    // Set cookie client-side immediately for instant theme switching
    document.cookie = `theme=${newTheme}; path=/; max-age=31536000; SameSite=Lax`

    // Update HTML class based on the new theme
    if (newTheme === 'dark') {
      document.documentElement.classList.remove('light')
      document.documentElement.classList.add('dark')
    } else if (newTheme === 'light') {
      document.documentElement.classList.remove('dark')
      document.documentElement.classList.add('light')
    } else {
      // System mode - detect current system preference
      const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      document.documentElement.classList.remove('light', 'dark')
      document.documentElement.classList.add(isDark ? 'dark' : 'light')
    }

    setCurrentTheme(newTheme)

    // Optionally call server action in background (non-blocking)
    // This ensures the cookie is also set server-side for future SSR
    if (setThemeCookie) {
      setThemeCookie(newTheme).catch(() => {
        // Silently ignore errors since client-side cookie is already set
      })
    }
  }

  const getIcon = () => {
    if (!mounted) return <Monitor className="h-4 w-4" />
    if (currentTheme === 'light') return <Sun className="h-4 w-4" />
    if (currentTheme === 'dark') return <Moon className="h-4 w-4" />
    return <Monitor className="h-4 w-4" />
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          className="inline-flex h-9 w-9 items-center justify-center rounded-md text-muted-foreground hover:bg-muted/50 hover:text-foreground transition-colors"
          aria-label="Toggle theme"
        >
          {getIcon()}
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onSelect={() => handleThemeChange('light')} className="cursor-pointer">
          <Sun className="mr-2 h-4 w-4" />
          <span>Light</span>
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => handleThemeChange('dark')} className="cursor-pointer">
          <Moon className="mr-2 h-4 w-4" />
          <span>Dark</span>
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => handleThemeChange('system')} className="cursor-pointer">
          <Monitor className="mr-2 h-4 w-4" />
          <span>System</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
