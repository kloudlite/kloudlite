import { getSession } from '@/lib/get-session'
import { redirect } from 'next/navigation'
import { signIn } from '@/lib/auth'
import { Shield, AlertCircle, AlertTriangle, LogIn } from 'lucide-react'
import {
  Card, CardContent, CardDescription, CardHeader, CardTitle,
  Alert, AlertDescription, Button,
} from '@kloudlite/ui'

export default async function SuperAdminLoginPage({
  searchParams,
}: {
  searchParams: Promise<{ token?: string; error?: string }>
}) {
  const { token, error } = await searchParams

  if (!token) {
    return <ErrorCard message="Missing authentication token" />
  }

  if (error) {
    return <ErrorCard message="Invalid or expired super-admin token" />
  }

  const session = await getSession()

  // If there's an existing session, show warning before proceeding
  if (session?.user) {
    return <SessionWarningCard token={token} session={session} />
  }

  // No existing session — server-side redirect to the API route handler which
  // can set session cookies (server components can only read cookies, not set them).
  // This is faster than the old <meta refresh> approach: a 302 vs rendering HTML.
  redirect(`/api/superadmin-login?token=${encodeURIComponent(token)}`)
}

function SessionWarningCard({ token, session }: { token: string; session: any }) {
  async function loginAsSuperAdmin() {
    'use server'
    await signIn('credentials', {
      superadminToken: token,
      redirectTo: '/admin',
    })
  }

  return (
    <div className="bg-background flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-amber-100 dark:bg-amber-900/30">
            <AlertTriangle className="h-8 w-8 text-amber-600 dark:text-amber-400" />
          </div>
          <CardTitle className="text-2xl">Active Session Detected</CardTitle>
          <CardDescription>
            You are currently logged in as <strong>{session.user.name || session.user.email}</strong>.
            Continuing will end that session.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-3">
            <form action={loginAsSuperAdmin}>
              <Button type="submit" className="w-full gap-2">
                <LogIn className="h-4 w-4" />
                Continue as Super Admin
              </Button>
            </form>
            <a href="/">
              <Button variant="outline" className="w-full">
                Cancel
              </Button>
            </a>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function ErrorCard({ message }: { message: string }) {
  return (
    <div className="bg-background flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="bg-primary/10 mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full">
            <Shield className="text-primary h-8 w-8" />
          </div>
          <CardTitle className="text-2xl">Super Admin Login</CardTitle>
          <CardDescription>Authentication failed</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{message}</AlertDescription>
            </Alert>
            <div className="text-muted-foreground text-center text-sm">
              <p>This login link may have expired or is invalid.</p>
              <p className="mt-2">Please generate a new login URL from the console.</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
