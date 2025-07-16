import { redirect } from 'next/navigation'
import { getSession } from '@/actions/auth/session'

export default async function TeamsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  // TEMPORARY: Bypass auth check for demo
  // const session = await getSession()
  
  // if (!session) {
  //   redirect('/auth/login')
  // }

  return <>{children}</>
}