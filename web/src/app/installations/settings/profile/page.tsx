import { redirect } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { getRegistrationSession } from '@/lib/registration-auth'
import { Separator } from '@/components/ui/separator'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { User, Mail, Shield, Info } from 'lucide-react'

export default async function ProfilePage() {
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/installations/login')
  }

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  return (
    <div className="space-y-6">
      {/* Profile Card */}
      <Card>
        <CardHeader>
          <CardTitle className="text-2xl">Profile Information</CardTitle>
          <CardDescription>Your account details from {session.provider} OAuth</CardDescription>
        </CardHeader>
        <CardContent className="space-y-8">
          {/* Profile Picture */}
          <div className="flex items-start gap-6">
            <Avatar className="h-24 w-24 ring-2 ring-border">
              <AvatarImage src={session.user.image} alt={session.user.name} />
              <AvatarFallback className="text-2xl">
                {getInitials(session.user.name)}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1 space-y-1 pt-2">
              <label className="text-sm font-semibold text-foreground">Profile Picture</label>
              <p className="text-sm text-muted-foreground">
                Synced from your {session.provider} account
              </p>
            </div>
          </div>

          <Separator />

          {/* Name */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <User className="h-4 w-4 text-muted-foreground" />
              <label className="text-sm font-semibold text-foreground">Name</label>
            </div>
            <div className="rounded-lg border border-border bg-muted/30 px-4 py-3">
              <p className="text-sm font-medium">{session.user.name}</p>
            </div>
            <p className="text-xs text-muted-foreground">
              Synced from your {session.provider} account
            </p>
          </div>

          <Separator />

          {/* Email */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Mail className="h-4 w-4 text-muted-foreground" />
              <label className="text-sm font-semibold text-foreground">Email Address</label>
            </div>
            <div className="rounded-lg border border-border bg-muted/30 px-4 py-3">
              <p className="text-sm font-medium">{session.user.email}</p>
            </div>
            <p className="text-xs text-muted-foreground">
              Primary email from your {session.provider} account
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Authentication Card */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Shield className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-xl">Authentication</CardTitle>
          </div>
          <CardDescription>Your authentication provider information</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            <label className="text-sm font-semibold text-foreground">OAuth Provider</label>
            <div className="flex items-center gap-3 rounded-lg border border-border bg-muted/30 px-4 py-3">
              <Badge variant="outline" className="capitalize">
                {session.provider}
              </Badge>
              <span className="text-sm text-muted-foreground">
                You're signed in with {session.provider}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Info Note - smaller and less prominent */}
      <div className="flex items-start gap-2 rounded-lg bg-muted/50 p-4">
        <Info className="h-4 w-4 text-muted-foreground mt-0.5 flex-shrink-0" />
        <p className="text-xs text-muted-foreground">
          Your profile information is managed by your OAuth provider ({session.provider}). To update your name, email, or profile picture, please update them in your {session.provider} account settings.
        </p>
      </div>
    </div>
  )
}
