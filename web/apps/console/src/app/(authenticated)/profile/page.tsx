import { getRegistrationSession } from '@/lib/console-auth'
import { Avatar, AvatarFallback, AvatarImage, Badge } from '@kloudlite/ui'
import { User, Mail, Shield, Info } from 'lucide-react'

export default async function ProfilePage() {
  const session = await getRegistrationSession()
  if (!session?.user) return null

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  return (
    <main className="mx-auto max-w-6xl px-6 lg:px-12 py-10">
      {/* Title */}
      <div className="mb-8">
        <h1 className="text-2xl font-semibold text-foreground">Profile</h1>
        <p className="text-muted-foreground mt-1 text-sm">
          Your personal account information
        </p>
      </div>

      {/* Profile Content */}
      <div className="space-y-6">
        {/* Profile Information Card */}
        <div className="border border-foreground/10 rounded-lg p-6 bg-background">
          <div className="mb-5">
            <h2 className="text-lg font-semibold">Profile Information</h2>
            <p className="text-muted-foreground mt-1 text-sm">Your account details from {session.provider} OAuth</p>
          </div>

          <div className="space-y-5">
            {/* Profile Picture */}
            <div className="flex items-start gap-6">
              <Avatar className="ring-foreground/10 h-24 w-24 ring-1">
                <AvatarImage src={session.user.image} alt={session.user.name} />
                <AvatarFallback className="text-2xl">{getInitials(session.user.name)}</AvatarFallback>
              </Avatar>
              <div className="flex-1 space-y-1 pt-2">
                <label className="text-foreground text-sm font-medium">Profile Picture</label>
                <p className="text-muted-foreground text-sm">
                  Synced from your {session.provider} account
                </p>
              </div>
            </div>

            <div className="h-px bg-foreground/10" />

            {/* Name */}
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <User className="text-muted-foreground h-4 w-4" />
                <label className="text-foreground text-sm font-medium">Name</label>
              </div>
              <div className="border-foreground/10 bg-muted/30 border px-4 py-3">
                <p className="text-sm">{session.user.name}</p>
              </div>
              <p className="text-muted-foreground text-sm">
                Synced from your {session.provider} account
              </p>
            </div>

            <div className="h-px bg-foreground/10" />

            {/* Email */}
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <Mail className="text-muted-foreground h-4 w-4" />
                <label className="text-foreground text-sm font-medium">Email Address</label>
              </div>
              <div className="border-foreground/10 bg-muted/30 border px-4 py-3">
                <p className="text-sm">{session.user.email}</p>
              </div>
              <p className="text-muted-foreground text-sm">
                Primary email from your {session.provider} account
              </p>
            </div>

            <div className="h-px bg-foreground/10" />

            {/* Authentication */}
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <Shield className="text-muted-foreground h-4 w-4" />
                <label className="text-foreground text-sm font-medium">Authentication Provider</label>
              </div>
              <div className="border-foreground/10 bg-muted/30 flex items-center gap-3 border px-4 py-3">
                <Badge variant="outline" className="capitalize">
                  {session.provider}
                </Badge>
                <span className="text-muted-foreground text-sm">
                  You&apos;re signed in with {session.provider}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Info Note */}
        <div className="bg-muted/30 flex items-start gap-3 border border-foreground/10 rounded-lg p-4">
          <Info className="text-muted-foreground mt-0.5 h-4 w-4 flex-shrink-0" />
          <p className="text-muted-foreground text-sm">
            Your profile information is managed by your OAuth provider ({session.provider}). To update
            your name, email, or profile picture, please update them in your {session.provider}{' '}
            account settings.
          </p>
        </div>
      </div>
    </main>
  )
}
