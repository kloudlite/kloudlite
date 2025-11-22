export type Theme = 'light' | 'dark'

const THEME_COOKIE_NAME = 'theme'

export function setThemeCookie(theme: Theme) {
  document.cookie = `${THEME_COOKIE_NAME}=${theme}; path=/; max-age=31536000; SameSite=Lax`
}

export function getThemeFromCookie(): Theme {
  if (typeof document === 'undefined') return 'light'

  const cookies = document.cookie.split('; ')
  const themeCookie = cookies.find((cookie) => cookie.startsWith(`${THEME_COOKIE_NAME}=`))
  return (themeCookie?.split('=')[1] as Theme) || 'light'
}
