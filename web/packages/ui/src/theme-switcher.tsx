'use client'

import { useEffect, useState } from 'react'
import { Moon, Sun, Monitor } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from './dropdown-menu'

type Theme = 'light' | 'dark'
type ThemeOption = Theme | 'system'

interface ThemeSwitcherProps {
  initialTheme?: ThemeOption
}

function setThemeCookie(theme: Theme | 'system') {
  document.cookie = `theme=${theme}; path=/; max-age=31536000; SameSite=Lax`
}

export function ThemeSwitcher({ initialTheme = 'light' }: ThemeSwitcherProps) {
  const [theme, setTheme] = useState<ThemeOption>(initialTheme)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  const applyTheme = (newTheme: ThemeOption) => {
    setTheme(newTheme)

    if (newTheme === 'system') {
      setThemeCookie('system')
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      document.documentElement.classList.toggle('dark', prefersDark)
    } else {
      setThemeCookie(newTheme)
      document.documentElement.classList.toggle('dark', newTheme === 'dark')
    }
  }

  const getIcon = () => {
    if (!mounted) return <Monitor className="h-4 w-4" />
    if (theme === 'light') return <Sun className="h-4 w-4" />
    if (theme === 'dark') return <Moon className="h-4 w-4" />
    return <Monitor className="h-4 w-4" />
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="text-muted-foreground hover:text-foreground transition-colors">
          {getIcon()}
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => applyTheme('light')}>
          <Sun className="mr-2 h-4 w-4" />
          <span>Light</span>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => applyTheme('dark')}>
          <Moon className="mr-2 h-4 w-4" />
          <span>Dark</span>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => applyTheme('system')}>
          <Monitor className="mr-2 h-4 w-4" />
          <span>System</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
