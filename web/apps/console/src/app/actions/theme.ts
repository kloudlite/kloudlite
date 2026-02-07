'use server'

import { cookies } from 'next/headers'

export async function setThemeCookie(theme: 'light' | 'dark' | 'system') {
  const cookieStore = await cookies()
  cookieStore.set('theme', theme, {
    path: '/',
    maxAge: 31536000, // 1 year
    sameSite: 'lax',
  })
}
