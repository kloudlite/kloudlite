'use client'

import * as React from 'react'

import { setThemeAction } from '@/app/actions/theme'
import { type Theme } from '@/lib/theme'

type ThemeProviderState = {
  theme: Theme
  setTheme: (theme: Theme) => void
}

const ThemeProviderContext = React.createContext<ThemeProviderState | undefined>(
  undefined
)

interface ThemeContextProviderProps {
  children: React.ReactNode
  defaultTheme: Theme
}

export function ThemeContextProvider({ children, defaultTheme }: ThemeContextProviderProps) {
  const [theme, setThemeState] = React.useState<Theme>(defaultTheme)

  React.useEffect(() => {
    const root = window.document.documentElement
    
    // Only update if theme actually changed
    if (theme === 'system') {
      const systemTheme = window.matchMedia('(prefers-color-scheme: dark)')
        .matches
        ? 'dark'
        : 'light'

      root.classList.remove('light', 'dark')
      root.classList.add(systemTheme)
      
      // Listen for system theme changes
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      const handleChange = (e: MediaQueryListEvent) => {
        if (theme === 'system') {
          root.classList.remove('light', 'dark')
          root.classList.add(e.matches ? 'dark' : 'light')
        }
      }
      mediaQuery.addEventListener('change', handleChange)
      
      return () => mediaQuery.removeEventListener('change', handleChange)
    }

    root.classList.remove('light', 'dark')
    root.classList.add(theme)
  }, [theme])

  const setTheme = React.useCallback(async (newTheme: Theme) => {
    setThemeState(newTheme)
    // Use server action to set cookie
    await setThemeAction(newTheme)
  }, [])

  const value = React.useMemo(
    () => ({
      theme,
      setTheme,
    }),
    [theme, setTheme]
  )

  return (
    <ThemeProviderContext.Provider value={value}>
      {children}
    </ThemeProviderContext.Provider>
  )
}

export const useTheme = () => {
  const context = React.useContext(ThemeProviderContext)

  if (context === undefined)
    {throw new Error('useTheme must be used within a ThemeProvider')}

  return context
}