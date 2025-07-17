'use client'

import { Moon, Sun } from 'lucide-react'
import { useEffect, useState } from 'react'

export function SimpleThemeToggle() {
  const [theme, setTheme] = useState<'light' | 'dark'>('light')
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
    // Get initial theme from document
    const currentTheme = document.documentElement.className as 'light' | 'dark'
    setTheme(currentTheme || 'light')
  }, [])

  const toggleTheme = () => {
    const newTheme = theme === 'light' ? 'dark' : 'light'
    setTheme(newTheme)
    document.documentElement.className = newTheme
    // Save to cookie
    document.cookie = `theme=${newTheme};path=/;max-age=${60 * 60 * 24 * 365};samesite=lax`
  }

  if (!mounted) {
    return (
      <button 
        className="size-8 rounded-md flex items-center justify-center bg-transparent hover:bg-muted transition-colors"
        disabled
      >
        <div className="size-4" />
      </button>
    )
  }

  return (
    <button
      onClick={toggleTheme}
      className="size-8 rounded-md flex items-center justify-center bg-transparent hover:bg-muted transition-colors"
      aria-label="Toggle theme"
    >
      {theme === 'light' ? (
        <Sun className="size-4 text-foreground" />
      ) : (
        <Moon className="size-4 text-foreground" />
      )}
    </button>
  )
}