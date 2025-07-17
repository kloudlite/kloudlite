'use client'

import { useState, useEffect } from 'react'
import { useSession, signIn, signOut } from 'next-auth/react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Copy, Check, LogOut, LogIn } from 'lucide-react'
import { cn } from '@/lib/utils'

export default function SSODemoPage() {
  const { data: session, status } = useSession()
  const [ssoEmail, setSsoEmail] = useState('')
  const [customDomain, setCustomDomain] = useState('')
  const [copied, setCopied] = useState(false)

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleSSOLogin = async () => {
    // For demo purposes, we'll use email domain to determine SSO provider
    const domain = ssoEmail.split('@')[1]
    
    // Mock SSO provider mapping
    const ssoProviders: Record<string, string> = {
      'microsoft.com': 'azure-ad',
      'google.com': 'google',
      'github.com': 'github',
      // Add custom domain mapping
      ...(customDomain ? { [customDomain]: 'azure-ad' } : {})
    }

    const provider = ssoProviders[domain]
    
    if (provider) {
      await signIn(provider, { callbackUrl: '/sso-demo' })
    } else {
      alert(`No SSO provider configured for domain: ${domain}`)
    }
  }

  return (
    <div className="min-h-screen bg-muted/30 p-6">
      <div className="max-w-4xl mx-auto space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>SSO Login Demo</CardTitle>
            <CardDescription>
              Test Single Sign-On functionality with different providers
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="status" className="w-full">
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="status">Status</TabsTrigger>
                <TabsTrigger value="login">SSO Login</TabsTrigger>
                <TabsTrigger value="config">Configuration</TabsTrigger>
              </TabsList>

              <TabsContent value="status" className="space-y-4">
                <div className="space-y-2">
                  <h3 className="text-lg font-semibold">Authentication Status</h3>
                  <div className={cn(
                    "inline-flex items-center gap-2 px-3 py-1 rounded-full text-sm font-medium",
                    status === 'authenticated' 
                      ? "bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400"
                      : status === 'loading'
                      ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400"
                      : "bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400"
                  )}>
                    <div className={cn(
                      "w-2 h-2 rounded-full",
                      status === 'authenticated' ? "bg-green-600" : 
                      status === 'loading' ? "bg-yellow-600" : "bg-red-600"
                    )} />
                    {status === 'authenticated' ? 'Authenticated' : 
                     status === 'loading' ? 'Loading...' : 'Not Authenticated'}
                  </div>
                </div>

                {session && (
                  <div className="space-y-4">
                    <div className="border rounded-lg p-4 bg-background">
                      <h4 className="font-medium mb-3">User Information</h4>
                      <dl className="space-y-2 text-sm">
                        <div className="flex justify-between">
                          <dt className="text-muted-foreground">Name:</dt>
                          <dd className="font-medium">{session.user?.name || 'N/A'}</dd>
                        </div>
                        <div className="flex justify-between">
                          <dt className="text-muted-foreground">Email:</dt>
                          <dd className="font-medium">{session.user?.email || 'N/A'}</dd>
                        </div>
                        <div className="flex justify-between">
                          <dt className="text-muted-foreground">Provider:</dt>
                          <dd className="font-medium">{session.provider || 'N/A'}</dd>
                        </div>
                      </dl>
                    </div>

                    <div className="border rounded-lg p-4 bg-background">
                      <h4 className="font-medium mb-3">Session Data</h4>
                      <pre className="text-xs bg-muted p-3 rounded overflow-x-auto">
                        {JSON.stringify(session, null, 2)}
                      </pre>
                      <Button
                        variant="outline"
                        size="sm"
                        className="mt-2"
                        onClick={() => handleCopy(JSON.stringify(session, null, 2))}
                      >
                        {copied ? <Check className="h-4 w-4 mr-2" /> : <Copy className="h-4 w-4 mr-2" />}
                        {copied ? 'Copied!' : 'Copy'}
                      </Button>
                    </div>

                    <Button
                      variant="destructive"
                      onClick={() => signOut({ callbackUrl: '/sso-demo' })}
                      className="w-full"
                    >
                      <LogOut className="h-4 w-4 mr-2" />
                      Sign Out
                    </Button>
                  </div>
                )}
              </TabsContent>

              <TabsContent value="login" className="space-y-4">
                <Alert>
                  <AlertDescription>
                    Enter your email address to automatically detect and use the appropriate SSO provider.
                  </AlertDescription>
                </Alert>

                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="sso-email">Email Address</Label>
                    <Input
                      id="sso-email"
                      type="email"
                      placeholder="user@company.com"
                      value={ssoEmail}
                      onChange={(e) => setSsoEmail(e.target.value)}
                    />
                    <p className="text-sm text-muted-foreground">
                      Supported domains: microsoft.com, google.com, github.com
                    </p>
                  </div>

                  <Button
                    onClick={handleSSOLogin}
                    disabled={!ssoEmail || status === 'loading'}
                    className="w-full"
                  >
                    <LogIn className="h-4 w-4 mr-2" />
                    Continue with SSO
                  </Button>

                  <div className="relative">
                    <div className="absolute inset-0 flex items-center">
                      <span className="w-full border-t" />
                    </div>
                    <div className="relative flex justify-center text-xs uppercase">
                      <span className="bg-background px-2 text-muted-foreground">
                        Or continue with
                      </span>
                    </div>
                  </div>

                  <div className="grid grid-cols-3 gap-3">
                    <Button
                      variant="outline"
                      onClick={() => signIn('google', { callbackUrl: '/sso-demo' })}
                      disabled={status === 'loading'}
                    >
                      <svg className="h-5 w-5" viewBox="0 0 24 24">
                        <path
                          fill="#4285F4"
                          d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                        />
                        <path
                          fill="#34A853"
                          d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                        />
                        <path
                          fill="#FBBC05"
                          d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                        />
                        <path
                          fill="#EA4335"
                          d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                        />
                      </svg>
                    </Button>

                    <Button
                      variant="outline"
                      onClick={() => signIn('azure-ad', { callbackUrl: '/sso-demo' })}
                      disabled={status === 'loading'}
                    >
                      <svg className="h-5 w-5" viewBox="0 0 23 23" fill="none">
                        <path d="M11 11H0V0H11V11Z" fill="#F25022"/>
                        <path d="M23 11H12V0H23V11Z" fill="#7FBA00"/>
                        <path d="M11 23H0V12H11V23Z" fill="#00A4EF"/>
                        <path d="M23 23H12V12H23V23Z" fill="#FFB900"/>
                      </svg>
                    </Button>

                    <Button
                      variant="outline"
                      onClick={() => signIn('github', { callbackUrl: '/sso-demo' })}
                      disabled={status === 'loading'}
                    >
                      <svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                      </svg>
                    </Button>
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="config" className="space-y-4">
                <Alert>
                  <AlertDescription>
                    Configure custom SSO domains for testing. In production, this would be managed by administrators.
                  </AlertDescription>
                </Alert>

                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="custom-domain">Custom Domain</Label>
                    <Input
                      id="custom-domain"
                      type="text"
                      placeholder="example.com"
                      value={customDomain}
                      onChange={(e) => setCustomDomain(e.target.value)}
                    />
                    <p className="text-sm text-muted-foreground">
                      Map a custom domain to an SSO provider (defaults to Azure AD)
                    </p>
                  </div>

                  <div className="border rounded-lg p-4 bg-background">
                    <h4 className="font-medium mb-3">Current SSO Mappings</h4>
                    <dl className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <dt className="text-muted-foreground">microsoft.com</dt>
                        <dd className="font-medium">Azure AD</dd>
                      </div>
                      <div className="flex justify-between">
                        <dt className="text-muted-foreground">google.com</dt>
                        <dd className="font-medium">Google</dd>
                      </div>
                      <div className="flex justify-between">
                        <dt className="text-muted-foreground">github.com</dt>
                        <dd className="font-medium">GitHub</dd>
                      </div>
                      {customDomain && (
                        <div className="flex justify-between">
                          <dt className="text-muted-foreground">{customDomain}</dt>
                          <dd className="font-medium">Azure AD (Custom)</dd>
                        </div>
                      )}
                    </dl>
                  </div>

                  <div className="border rounded-lg p-4 bg-background">
                    <h4 className="font-medium mb-3">Environment Variables</h4>
                    <pre className="text-xs bg-muted p-3 rounded overflow-x-auto">
{`NEXTAUTH_URL=${process.env.NEXTAUTH_URL || 'http://localhost:3000'}
NEXTAUTH_SECRET=*****

# Google OAuth
GOOGLE_CLIENT_ID=${process.env.GOOGLE_CLIENT_ID ? '✓ Configured' : '✗ Not set'}
GOOGLE_CLIENT_SECRET=${process.env.GOOGLE_CLIENT_SECRET ? '✓ Configured' : '✗ Not set'}

# Azure AD OAuth
AZURE_AD_CLIENT_ID=${process.env.AZURE_AD_CLIENT_ID ? '✓ Configured' : '✗ Not set'}
AZURE_AD_CLIENT_SECRET=${process.env.AZURE_AD_CLIENT_SECRET ? '✓ Configured' : '✗ Not set'}
AZURE_AD_TENANT_ID=${process.env.AZURE_AD_TENANT_ID ? '✓ Configured' : '✗ Not set'}

# GitHub OAuth
GITHUB_CLIENT_ID=${process.env.GITHUB_CLIENT_ID ? '✓ Configured' : '✗ Not set'}
GITHUB_CLIENT_SECRET=${process.env.GITHUB_CLIENT_SECRET ? '✓ Configured' : '✗ Not set'}`}
                    </pre>
                  </div>
                </div>
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}