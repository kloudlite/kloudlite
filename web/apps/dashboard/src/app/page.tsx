import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'

export default async function HomePage() {
  const session = await auth()

  if (session) {
    // User is logged in, redirect to workspaces
    redirect('/workspaces')
  } else {
    // User is not logged in, redirect to sign in
    redirect('/auth/signin')
  }
}
