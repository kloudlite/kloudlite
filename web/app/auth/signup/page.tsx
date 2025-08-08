import Link from "next/link"
import { redirect } from "next/navigation"
import { getServerSession } from "next-auth"

import { SignupForm } from "@/components/auth/signup-form"
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
  searchParams: Promise<{
    callbackUrl?: string
  }>
}

export default async function SignupPage({ searchParams }: PageProps) {
  const session = await getServerSession(authOptions)
  const params = await searchParams
  const callbackUrl = params.callbackUrl || "/overview"
  
  // Redirect if already logged in
  if (session) {
    redirect(callbackUrl)
  }

  // Get platform settings
  let platformSettings = null
  try {
    const client = getAccountsClient()
    const metadata = new (await import('@grpc/grpc-js')).Metadata()
    const response = await new Promise<any>((resolve, reject) => {
      client.getPlatformSettings({}, metadata, (error, response) => {
        if (error) {reject(error)}
        else {resolve(response)}
      })
    })
    platformSettings = response?.settings
  } catch (error) {
    console.error('Failed to fetch platform settings:', error)
  }

  // If signup is not allowed, redirect to login
  if (platformSettings && !platformSettings.allowSignup) {
    redirect("/auth/login")
  }

  return (
    <Card>
      <CardHeader className="text-center">
        <CardTitle className="text-xl">Create an account</CardTitle>
        <CardDescription>
          Enter your information to get started
        </CardDescription>
      </CardHeader>
      <CardContent>
        <SignupForm callbackUrl={callbackUrl} platformSettings={platformSettings} />
        
        <div className="mt-6 text-center text-sm">
          Already have an account?{" "}
          <Link href="/auth/login" className="underline underline-offset-4">
            Sign in
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}