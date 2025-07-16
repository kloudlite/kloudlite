import { redirect } from 'next/navigation'
import { getSession } from '@/actions/auth/session'
import { Link } from '@/components/ui/link'
import { Button } from '@/components/ui/button'
import { Users, LogOut } from 'lucide-react'
import { logoutAction } from '@/actions/auth/logout'

export default async function DashboardPage() {
  const session = await getSession()
  
  if (!session) {
    redirect('/auth/login')
  }

  async function logout() {
    'use server'
    await logoutAction()
    redirect('/auth/login')
  }

  return (
    <div className="container mx-auto py-10">
      <div className="flex justify-between items-start mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-4">Dashboard</h1>
          <p className="text-muted-foreground">Welcome back, {session.user.name}!</p>
          <p className="text-sm text-muted-foreground mt-2">Email: {session.user.email}</p>
        </div>
        <form action={logout}>
          <Button type="submit" variant="outline" className="rounded-none">
            <LogOut className="h-4 w-4 mr-2" />
            Logout
          </Button>
        </form>
      </div>
      
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 mt-8">
        <Link href="/teams" className="block">
          <div className="border border-border rounded-none p-6 hover:shadow-sm transition-shadow">
            <Users className="h-8 w-8 mb-4 text-primary" />
            <h2 className="text-xl font-semibold mb-2">Teams</h2>
            <p className="text-muted-foreground">Manage your teams and collaborate with others</p>
          </div>
        </Link>
      </div>
    </div>
  )
}