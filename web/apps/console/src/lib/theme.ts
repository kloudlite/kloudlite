export type Theme = 'light' | 'dark'

const THEME_COOKIE_NAME = 'theme'

export function setThemeCookie(theme: Theme) {
  document.cookie = `${THEME_COOKIE_NAME}=${theme}; path=/; max-age=31536000; SameSite=Lax`
}
