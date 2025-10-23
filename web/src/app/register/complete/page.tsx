'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ExternalLink, Loader2, Copy, Trash2 } from 'lucide-react'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { toast } from 'sonner'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface UserData {
  subdomain: string
  url: string
  installationKey: string
}

const getUninstallCommands = (installationKey: string) => ({
  aws: {
    name: 'AWS',
    commands: [
      `curl -fsSL https://get.kloudlite.io/aws | bash -s -- --uninstall --key ${installationKey}`,
    ],
  },
  gcp: {
    name: 'Google Cloud',
    commands: [
      `curl -fsSL https://get.kloudlite.io/gcp | bash -s -- --uninstall --key ${installationKey}`,
    ],
  },
  azure: {
    name: 'Azure',
    commands: [
      `curl -fsSL https://get.kloudlite.io/azure | bash -s -- --uninstall --key ${installationKey}`,
    ],
  },
})

export default function CompletePage() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [userData, setUserData] = useState<UserData | null>(null)
  const [selectedProvider, setSelectedProvider] = useState('aws')

  useEffect(() => {
    // Middleware handles all redirects based on state
    const fetchUserData = async () => {
      try {
        const response = await fetch('/api/register/session')
        if (response.ok) {
          const sessionData = await response.json()

          // Fetch full user data with installation key
          if (sessionData.installationKey) {
            const verifyResponse = await fetch('/api/register/verify-key', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ installationKey: sessionData.installationKey }),
            })
            if (verifyResponse.ok) {
              const verifyData = await verifyResponse.json()
              // Add installation key from session to user data
              setUserData({
                ...verifyData.user,
                installationKey: sessionData.installationKey,
              })
            }
          }
        }
      } catch (error) {
        console.error('Error fetching user data:', error)
      } finally {
        setLoading(false)
      }
    }
    fetchUserData()
  }, [router])

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} copied to clipboard`)
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="size-8 animate-spin text-primary" />
      </div>
    )
  }

  if (!userData || !userData.subdomain) {
    return null
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-8 bg-background">
      <div className="w-full max-w-2xl">
        <div className="mb-8 text-center">
          <KloudliteLogo className="mx-auto mb-6" />
        </div>

        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Installation Details</CardTitle>
              <CardDescription>
                Your Kloudlite dashboard is ready
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="bg-muted p-4 rounded-lg">
                <p className="text-sm font-medium mb-2">Dashboard URL:</p>
                <div className="flex items-center justify-between gap-2">
                  <a
                    href={userData.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline font-mono text-sm flex items-center gap-1"
                  >
                    {userData.subdomain}.kloudlite.io
                    <ExternalLink className="size-3" />
                  </a>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0"
                    onClick={() => copyToClipboard(userData.url, 'URL')}
                  >
                    <Copy className="size-3" />
                  </Button>
                </div>
              </div>

              <Button
                className="w-full"
                onClick={() => window.open(userData.url, '_blank')}
              >
                Open Dashboard
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Trash2 className="size-4 text-destructive" />
                Uninstall Kloudlite
              </CardTitle>
              <CardDescription>
                Remove Kloudlite from your cloud provider
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <Tabs value={selectedProvider} onValueChange={setSelectedProvider}>
                <TabsList className="grid w-full grid-cols-3">
                  <TabsTrigger value="aws">AWS</TabsTrigger>
                  <TabsTrigger value="gcp">GCP</TabsTrigger>
                  <TabsTrigger value="azure">Azure</TabsTrigger>
                </TabsList>

                {Object.entries(getUninstallCommands(userData.installationKey)).map(([key, config]) => (
                  <TabsContent key={key} value={key} className="space-y-3 mt-4">
                    <div className="flex items-center gap-2 text-sm font-medium">
                      <Trash2 className="size-4" />
                      Uninstall from {config.name}
                    </div>

                    <div className="space-y-2">
                      {config.commands.map((cmd, idx) => (
                        <div key={idx} className="bg-muted rounded-lg p-3">
                          <div className="flex items-start justify-between gap-2">
                            <code className="text-xs font-mono flex-1 break-all">{cmd}</code>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-6 w-6 p-0 flex-shrink-0"
                              onClick={() => copyToClipboard(cmd, 'Command')}
                            >
                              <Copy className="size-3" />
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>

                    <p className="text-xs text-muted-foreground">
                      This will remove all Kloudlite resources from your {config.name} account
                    </p>
                  </TabsContent>
                ))}
              </Tabs>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
