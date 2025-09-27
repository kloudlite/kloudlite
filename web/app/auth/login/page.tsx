import Link from "next/link"
import { redirect } from "next/navigation"
import { getServerSession } from "next-auth"

import { LoginForm } from "@/components/auth/login-form"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { authOptions } from "@/lib/auth/get-auth-options"
import { getAccountsClient } from "@/lib/grpc/accounts-client"

interface PageProps {
  searchParams: {
    callbackUrl?: string
  }
}

export default async function LoginPage({ searchParams }: PageProps) {
  const session = await getServerSession(authOptions)
  const callbackUrl = searchParams.callbackUrl || "/overview"
  
  // Redirect if already logged in
  if (session) {
    redirect(callbackUrl)
  }

  // Get platform settings
  let platformSettings = null
  try {
    const client = getAccountsClient()
    const response = await new Promise<any>((resolve, reject) => {
      client.getPlatformSettings({}, {}, (error, response) => {
        if (error) {reject(error)}
        else {resolve(response)}
      })
    })
    platformSettings = response?.settings
  } catch (error) {
    console.error('Failed to fetch platform settings:', error)
  }

  return (
    <Card>
      <CardHeader className="text-center">
        <CardTitle className="text-xl">Welcome back</CardTitle>
        <CardDescription>
          Sign in to your account to continue
        </CardDescription>
      </CardHeader>
      <CardContent>
        <LoginForm callbackUrl={callbackUrl} platformSettings={platformSettings} />
        
        {platformSettings?.allowSignup && (
          <div className="mt-6 text-center text-sm">
            Don&apos;t have an account?{" "}
            <Link href="/auth/signup" className="underline underline-offset-4">
              Sign up
            </Link>
          </div>
        )}
      </CardContent>
    </Card>
  )
}