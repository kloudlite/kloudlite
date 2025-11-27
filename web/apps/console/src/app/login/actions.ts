'use server'

import { signIn } from '@/lib/auth'

export async function handleOAuthSignIn(provider: string) {
  await signIn(provider, { redirectTo: '/access-console/install' })
}
