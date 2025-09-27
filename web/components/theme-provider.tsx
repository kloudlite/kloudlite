import { type Theme } from '@/lib/theme'

import { ThemeContextProvider, useTheme } from './theme-context'

export { useTheme }

type ThemeProviderProps = {
  children: React.ReactNode
  defaultTheme?: Theme
}

export function ThemeProvider({
  children,
  defaultTheme = 'system',
}: ThemeProviderProps) {
  return (
    <ThemeContextProvider defaultTheme={defaultTheme}>
      {children}
    </ThemeContextProvider>
  )
}