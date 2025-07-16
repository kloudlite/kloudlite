'use server'

import { cookies } from 'next/headers'
import { clearSessionCookie } from './session'

export async function logoutAction() {
  await clearSessionCookie()
}