import { getTheme } from '@/lib/theme-server'
import { ThemeSwitcher } from './theme-switcher'

export async function ThemeSwitcherServer() {
  const theme = await getTheme()
  return <ThemeSwitcher initialTheme={theme} />
}
