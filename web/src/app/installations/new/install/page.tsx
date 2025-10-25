'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Loader2, Cloud, Copy, CheckCircle2, Clock } from 'lucide-react'
import { InstallationProgress } from '@/components/installation-progress'
import { toast } from 'sonner'

interface SessionData {
  user: {
    email: string
    name: string
  }
  installationKey: string
}

const getCloudProviderCommands = (installationKey: string) => ({
  aws: {
    name: 'AWS',
    commands: [`curl -fsSL https://get.kloudlite.io/aws | bash -s -- --key ${installationKey}`],
    requirements: [
      'AWS CLI configured',
      'IAM user with EC2 full access and iam:PassRole permissions',
      'Valid AWS access keys configured',
    ],
  },
  gcp: {
    name: 'Google Cloud',
    commands: [`curl -fsSL https://get.kloudlite.io/gcp | bash -s -- --key ${installationKey}`],
    requirements: [
      'gcloud CLI configured',
      'Service account with Compute Admin and Service Account User roles',
      'Valid GCP credentials configured',
    ],
  },
  azure: {
    name: 'Azure',
    commands: [`curl -fsSL https://get.kloudlite.io/azure | bash -s -- --key ${installationKey}`],
    requirements: [
      'Azure CLI configured',
      'Service principal with Virtual Machine Contributor and User Access Administrator roles',
      'Valid Azure credentials configured',
    ],
  },
})

export default function InstallPage() {
  const router = useRouter()
  const [selectedProvider, setSelectedProvider] = useState('aws')
  const [session, setSession] = useState<SessionData | null>(null)
  const [loading, setLoading] = useState(true)
  const [verificationStatus, setVerificationStatus] = useState<'waiting' | 'verified' | 'error'>(
    'waiting',
  )

  // Check session on mount
  useEffect(() => {
    const checkSession = async () => {
      try {
        const response = await fetch('/api/installations/session')
        if (response.ok) {
          const data = await response.json()

          if (!data.installationKey) {
            // No installation key, redirect to create installation page
            console.log('No installationKey in session, redirecting to /installations/new')
            router.push('/installations/new')
            return
          }

          setSession(data)
        } else {
          console.log('Session check failed, redirecting to /installations/login')
          router.push('/installations/login')
        }
      } catch (error) {
        console.error('Error checking session:', error)
        router.push('/installations/login')
      } finally {
        setLoading(false)
      }
    }
    checkSession()
  }, [router])

  // Poll for deployment verification every 5 seconds
  useEffect(() => {
    if (!session?.installationKey) return

    const checkVerification = async () => {
      try {
        const response = await fetch('/api/installations/check-installation', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ installationKey: session.installationKey }),
        })

        const data = await response.json()

        if (data.verified) {
          setVerificationStatus('verified')
          // Auto-redirect to domain page after 2 seconds
          setTimeout(() => {
            router.push('/installations/new/domain')
          }, 2000)
        }
      } catch (error) {
        console.error('Error checking verification:', error)
      }
    }

    // Check immediately
    checkVerification()

    // Then poll every 5 seconds
    const interval = setInterval(checkVerification, 5000)

    return () => clearInterval(interval)
  }, [session, router])

  // Redirect if no session after loading completes
  useEffect(() => {
    if (!loading && (!session || !session.installationKey)) {
      console.log('No session or installationKey, redirecting')
      router.push('/installations/new')
    }
  }, [loading, session, router])

  const copyCommand = (command: string) => {
    navigator.clipboard.writeText(command)
    toast.success('Command copied to clipboard')
  }

  if (loading) {
    return (
      <div className="bg-background flex min-h-screen items-center justify-center">
        <Loader2 className="text-primary size-8 animate-spin" />
      </div>
    )
  }

  if (!session || !session.installationKey) {
    return null
  }

  const CLOUD_PROVIDERS = getCloudProviderCommands(session.installationKey)

  return (
    <>
      {/* Header */}
      <div className="mb-12 text-center">
        <h1 className="text-foreground mb-3 text-4xl font-bold tracking-tight">
          Install Kloudlite in Your Cloud
        </h1>
        <p className="text-muted-foreground text-lg">
          Run the installation command on your cloud provider
        </p>
      </div>

      <InstallationProgress currentStep={2} />

      {/* Verification Status Card */}
      <Card className="mb-8">
        <CardContent className="py-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              {verificationStatus === 'waiting' && (
                <>
                  <div className="flex size-12 items-center justify-center rounded-full bg-blue-50">
                    <Loader2 className="size-6 animate-spin text-blue-600" />
                  </div>
                  <div>
                    <p className="text-base font-semibold">Waiting for deployment...</p>
                    <p className="text-muted-foreground mt-0.5 text-sm">
                      Run the command below to start installation
                    </p>
                  </div>
                </>
              )}
              {verificationStatus === 'verified' && (
                <>
                  <div className="flex size-12 items-center justify-center rounded-full bg-green-50">
                    <CheckCircle2 className="size-6 text-green-600" />
                  </div>
                  <div>
                    <p className="text-base font-semibold text-green-600">Installation verified!</p>
                    <p className="text-muted-foreground mt-0.5 text-sm">
                      Redirecting to domain configuration...
                    </p>
                  </div>
                </>
              )}
            </div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Clock className="size-4" />
              <span className="font-medium">Checking status...</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Installation Commands Card */}
      <Card>
        <CardHeader className="pb-6">
          <CardTitle className="text-2xl font-bold">Installation Command</CardTitle>
          <CardDescription className="text-base">
            Choose your cloud provider and run the command in your terminal
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <Tabs value={selectedProvider} onValueChange={setSelectedProvider}>
            <TabsList className="grid w-full grid-cols-3 p-1">
              <TabsTrigger value="aws" className="text-sm font-semibold">AWS</TabsTrigger>
              <TabsTrigger value="gcp" className="text-sm font-semibold">GCP</TabsTrigger>
              <TabsTrigger value="azure" className="text-sm font-semibold">Azure</TabsTrigger>
            </TabsList>

            {Object.entries(CLOUD_PROVIDERS).map(([key, config]) => (
              <TabsContent key={key} value={key} className="mt-6 space-y-6">
                <div className="flex items-center gap-2 text-base font-semibold">
                  <Cloud className="size-5" />
                  Installing on {config.name}
                </div>

                <div className="space-y-5">
                  <div>
                    <p className="mb-3 text-sm font-semibold text-foreground">Prerequisites:</p>
                    <ul className="text-muted-foreground space-y-2 text-sm leading-relaxed">
                      {config.requirements.map((req, idx) => (
                        <li key={idx} className="flex items-start gap-3">
                          <div className="bg-muted-foreground mt-2 size-1.5 flex-shrink-0 rounded-full" />
                          <span>{req}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  <div>
                    <p className="mb-3 text-sm font-semibold text-foreground">Run this command:</p>
                    <div className="space-y-3">
                      {config.commands.map((cmd, idx) => (
                        <div key={idx} className="bg-muted rounded-lg p-4">
                          <div className="flex items-start justify-between gap-4">
                            <code className="flex-1 break-all font-mono text-sm leading-relaxed">{cmd}</code>
                            <Button
                              variant="outline"
                              size="sm"
                              className="flex-shrink-0"
                              onClick={() => copyCommand(cmd)}
                            >
                              <Copy className="mr-2 size-4" />
                              Copy
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="rounded-lg border border-blue-200 bg-blue-50 p-4">
                    <p className="text-sm leading-relaxed text-blue-900">
                      <span className="font-semibold">Note:</span> After running the command, this page will automatically
                      detect when your deployment contacts our server and proceed to the next step.
                    </p>
                  </div>
                </div>
              </TabsContent>
            ))}
          </Tabs>
        </CardContent>
      </Card>
    </>
  )
}
