'use client'

import { useEffect, useState } from 'react'
import { Moon, Sun } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { setThemeCookie, getThemeFromCookie, type Theme } from '@/lib/theme'

export function ThemeSwitcher() {
  const [theme, setTheme] = useState<Theme>(() => getThemeFromCookie())

  useEffect(() => {
    // Sync state with cookie on mount
    setTheme(getThemeFromCookie())
  }, [])

  const toggleTheme = () => {
    const newTheme: Theme = theme === 'light' ? 'dark' : 'light'
    setTheme(newTheme)
    setThemeCookie(newTheme)
    document.documentElement.classList.toggle('dark', newTheme === 'dark')
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={toggleTheme}
      aria-label="Toggle theme"
    >
      {/* Show moon icon in light mode, sun icon in dark mode */}
      <Moon className="h-4 w-4 dark:hidden" />
      <Sun className="h-4 w-4 hidden dark:block" />
    </Button>
  )
}
