import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'

export default async function HomePage() {
  const session = await getRegistrationSession()

  if (session?.user) {
    // User is logged in, redirect to installations
    redirect('/installations')
  } else {
    // User is not logged in, redirect to login
    redirect('/login')
  }
}
