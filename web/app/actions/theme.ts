'use server'

import { cookies } from 'next/headers'

import { type Theme } from '@/lib/theme'

export async function setThemeAction(theme: Theme) {
  const cookieStore = await cookies()
  
  cookieStore.set('theme', theme, {
    path: '/',
    maxAge: 60 * 60 * 24 * 365, // 1 year
    sameSite: 'lax',
  })
}