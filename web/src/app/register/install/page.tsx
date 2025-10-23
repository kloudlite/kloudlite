'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Loader2, Cloud, Copy, LogOut } from 'lucide-react'
import { KloudliteLogo } from '@/components/kloudlite-logo'
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

  useEffect(() => {
    // Check for registration session cookie
    const checkSession = async () => {
      try {
        const response = await fetch('/api/register/session')
        if (response.ok) {
          const data = await response.json()
          setSession(data)

          // Middleware handles all redirects - just set session
        } else {
          router.push('/register')
        }
      } catch {
        router.push('/register')
      } finally {
        setLoading(false)
      }
    }
    checkSession()
  }, [router])

  const copyCommand = (command: string) => {
    navigator.clipboard.writeText(command)
    toast.success('Command copied to clipboard')
  }

  const handleContinue = () => {
    // Middleware will handle redirect based on actual deployment state
    // Just navigate to next step - middleware will redirect if needed
    router.push('/register/domain')
  }

  const handleSignOut = async () => {
    try {
      // Clear the registration session cookie
      await fetch('/api/register/signout', { method: 'POST' })
      toast.success('Signed out successfully')
      router.push('/register')
    } catch {
      toast.error('Failed to sign out')
    }
  }

  if (loading) {
    return (
      <div className="bg-background flex min-h-screen items-center justify-center">
        <Loader2 className="text-primary size-8 animate-spin" />
      </div>
    )
  }

  if (!session) {
    return null
  }

  // If no installation key, redirect to re-authenticate
  if (!session.installationKey) {
    toast.error('Please sign in again to generate your installation key')
    handleSignOut()
    return null
  }

  const CLOUD_PROVIDERS = getCloudProviderCommands(session.installationKey)

  return (
    <div className="bg-background flex min-h-screen items-center justify-center p-8">
      <div className="w-full max-w-3xl">
        <div className="mb-8 flex flex-col items-center text-center">
          <KloudliteLogo className="mb-6 h-8" />
          <h1 className="text-foreground mb-2 text-2xl font-medium">
            Welcome, {session.user?.name || 'User'}!
          </h1>
          <p className="text-muted-foreground mb-4 text-sm">{session.user?.email || ''}</p>
          <div className="bg-muted inline-flex items-center gap-2 rounded-md px-3 py-1.5">
            <span className="text-muted-foreground text-xs">Installation Key:</span>
            <code className="text-foreground font-mono text-xs">{session.installationKey}</code>
            <Button
              variant="ghost"
              size="sm"
              className="h-5 w-5 p-0"
              onClick={() => {
                navigator.clipboard.writeText(session.installationKey)
                toast.success('Installation key copied')
              }}
            >
              <Copy className="size-3" />
            </Button>
          </div>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="text-xl">Install Kloudlite</CardTitle>
            <CardDescription>
              Choose your cloud provider and follow the installation steps
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <Tabs value={selectedProvider} onValueChange={setSelectedProvider}>
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="aws">AWS</TabsTrigger>
                <TabsTrigger value="gcp">GCP</TabsTrigger>
                <TabsTrigger value="azure">Azure</TabsTrigger>
              </TabsList>

              {Object.entries(CLOUD_PROVIDERS).map(([key, config]) => (
                <TabsContent key={key} value={key} className="mt-4 space-y-4">
                  <div className="flex items-center gap-2 text-sm font-medium">
                    <Cloud className="size-4" />
                    Installing on {config.name}
                  </div>

                  <div className="space-y-3">
                    <div>
                      <p className="mb-2 text-sm font-medium">Prerequisites:</p>
                      <ul className="text-muted-foreground space-y-1 text-sm">
                        {config.requirements.map((req, idx) => (
                          <li key={idx} className="flex items-center gap-2">
                            <div className="bg-muted-foreground size-1.5 rounded-full" />
                            {req}
                          </li>
                        ))}
                      </ul>
                    </div>

                    <div>
                      <p className="mb-2 text-sm font-medium">Installation steps:</p>
                      <div className="space-y-2">
                        {config.commands.map((cmd, idx) => (
                          <div key={idx} className="bg-muted rounded-lg p-3">
                            <div className="flex items-start justify-between gap-2">
                              <code className="flex-1 font-mono text-xs">{cmd}</code>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 w-6 p-0"
                                onClick={() => copyCommand(cmd)}
                              >
                                <Copy className="size-3" />
                              </Button>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                </TabsContent>
              ))}
            </Tabs>

            <div className="border-t pt-4">
              <Button onClick={handleContinue} className="w-full">
                I&apos;ve completed installation
              </Button>
              <p className="text-muted-foreground mt-2 text-center text-xs">
                Click continue after running the installation commands
              </p>
            </div>
          </CardContent>
        </Card>

        <div className="mt-4">
          <Button variant="ghost" size="sm" onClick={handleSignOut} className="gap-2">
            <LogOut className="size-4" />
            Sign out
          </Button>
        </div>
      </div>
    </div>
  )
}
