import { cookies } from 'next/headers'

export type Theme = 'light' | 'dark'

const THEME_COOKIE_NAME = 'theme'

export async function getTheme(): Promise<Theme> {
  const cookieStore = await cookies()
  const themeCookie = cookieStore.get(THEME_COOKIE_NAME)
  return (themeCookie?.value as Theme) || 'light'
}
