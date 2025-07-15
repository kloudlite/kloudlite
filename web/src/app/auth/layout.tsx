import { redirect } from 'next/navigation'
import { getSession } from '@/actions/auth/session'

export default async function AuthLayout({
  children,
}: {
  children: React.ReactNode
}) {
  // Check if user is already authenticated
  const session = await getSession()
  
  if (session) {
    redirect('/dashboard')
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {children}
      </div>
    </div>
  )
}