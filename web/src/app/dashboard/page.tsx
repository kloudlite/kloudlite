import { redirect } from 'next/navigation'
import { getSession } from '@/actions/auth/session'

export default async function DashboardPage() {
  const session = await getSession()
  
  if (!session) {
    redirect('/auth/login')
  }

  return (
    <div className="container mx-auto py-10">
      <h1 className="text-3xl font-bold mb-4">Dashboard</h1>
      <p className="text-muted-foreground">Welcome back, {session.user.name}!</p>
      <p className="text-sm text-muted-foreground mt-2">Email: {session.user.email}</p>
    </div>
  )
}