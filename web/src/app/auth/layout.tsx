import { redirect } from 'next/navigation'
import { getSession } from '@/actions/auth/session'
import { Link } from '@/components/ui/link'
import { Button } from '@/components/ui/button'
import { ArrowLeft } from 'lucide-react'

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
    <div className="min-h-screen flex items-center justify-center p-4 relative">
      <div className="absolute top-4 left-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Home
          </Link>
        </Button>
      </div>
      <div className="w-full max-w-md">
        {children}
      </div>
    </div>
  )
}