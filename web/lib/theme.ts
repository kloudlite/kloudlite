export type Theme = 'light' | 'dark' | 'system'

export function getThemeFromCookie(cookie: string | undefined): Theme {
  if (!cookie) {return 'system'}
  const match = cookie.match(/theme=(light|dark|system)/)
  return (match?.[1] as Theme) || 'system'
}

export function setThemeCookie(theme: Theme): string {
  return `theme=${theme}; path=/; max-age=${60 * 60 * 24 * 365}; SameSite=Lax`
}