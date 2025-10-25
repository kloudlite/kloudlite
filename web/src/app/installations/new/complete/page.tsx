'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ExternalLink, Loader2, Copy, PartyPopper } from 'lucide-react'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { InstallationProgress } from '@/components/installation-progress'
import { toast } from 'sonner'

interface InstallationData {
  subdomain: string
  url: string
}

export default function CompletePage() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [installationData, setInstallationData] = useState<InstallationData | null>(null)

  useEffect(() => {
    const fetchInstallationData = async () => {
      try {
        const response = await fetch('/api/installations/session')
        if (response.ok) {
          const sessionData = await response.json()

          // Fetch installation details using verify-key
          if (sessionData.installationKey) {
            const verifyResponse = await fetch('/api/installations/verify-key', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ installationKey: sessionData.installationKey }),
            })
            if (verifyResponse.ok) {
              const verifyData = await verifyResponse.json()
              if (verifyData.subdomain) {
                const domain = process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'
                setInstallationData({
                  subdomain: verifyData.subdomain,
                  url: `https://${verifyData.subdomain}.${domain}`,
                })
              }
            }
          }
        } else {
          router.push('/installations/login')
        }
      } catch (error) {
        console.error('Error fetching installation data:', error)
      } finally {
        setLoading(false)
      }
    }
    fetchInstallationData()
  }, [router])

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} copied to clipboard`)
  }

  if (loading) {
    return (
      <div className="bg-background flex min-h-screen items-center justify-center">
        <Loader2 className="text-primary size-8 animate-spin" />
      </div>
    )
  }

  if (!installationData || !installationData.subdomain) {
    return null
  }

  return (
    <div className="bg-background flex min-h-screen items-center justify-center p-8">
      <div className="w-full max-w-2xl">
        <div className="mb-8 text-center">
          <KloudliteLogo className="mx-auto mb-6" />
          <div className="mb-4 flex justify-center">
            <div className="flex size-16 items-center justify-center rounded-full bg-green-100">
              <PartyPopper className="size-8 text-green-600" />
            </div>
          </div>
          <h1 className="text-foreground mb-2 text-3xl font-semibold">Installation Complete!</h1>
          <p className="text-muted-foreground">
            Your Kloudlite installation is ready to use
          </p>
        </div>

        <InstallationProgress currentStep={4} />

        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-xl">Your Installation is Ready</CardTitle>
              <CardDescription>
                Access your Kloudlite installation dashboard at the URL below
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="bg-muted rounded-lg p-4">
                <p className="mb-2 text-sm font-medium">Installation Dashboard URL:</p>
                <div className="flex items-center justify-between gap-3">
                  <a
                    href={installationData.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary flex items-center gap-2 font-mono text-lg hover:underline"
                  >
                    {installationData.subdomain}.{process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                    <ExternalLink className="size-4" />
                  </a>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => copyToClipboard(installationData.url, 'URL')}
                  >
                    <Copy className="mr-2 size-3" />
                    Copy
                  </Button>
                </div>
              </div>

              <div className="flex gap-3">
                <Button
                  className="flex-1"
                  size="lg"
                  onClick={() => window.open(installationData.url, '_blank')}
                >
                  <ExternalLink className="mr-2 size-4" />
                  Open Installation Dashboard
                </Button>
                <Button
                  variant="outline"
                  size="lg"
                  className="flex-1"
                  onClick={() => router.push('/installations')}
                >
                  View All Installations
                </Button>
              </div>

              <div className="border-t pt-4">
                <p className="text-muted-foreground text-sm">
                  <strong>What's next?</strong> You can now access your Kloudlite installation dashboard to
                  create and manage workspaces, environments, and work machines. Your team members can log in
                  using their own credentials at your installation URL.
                </p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Need Help?</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <a
                href="https://docs.kloudlite.io"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary flex items-center gap-2 text-sm hover:underline"
              >
                <ExternalLink className="size-4" />
                Read the Documentation
              </a>
              <a
                href="https://discord.gg/kloudlite"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary flex items-center gap-2 text-sm hover:underline"
              >
                <ExternalLink className="size-4" />
                Join our Discord Community
              </a>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
