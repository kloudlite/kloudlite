import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { Separator } from '@kloudlite/ui'
import { Avatar, AvatarFallback, AvatarImage } from '@kloudlite/ui'
import { Badge } from '@kloudlite/ui'
import { User, Mail, Shield, Info } from 'lucide-react'

export default async function ProfilePage() {
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
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
    <div className="space-y-8">
      {/* Profile Information */}
      <div className="border border-foreground/10 p-8">
        <div className="mb-8">
          <h2 className="text-2xl font-semibold">Profile Information</h2>
          <p className="text-muted-foreground mt-1 text-base">Your account details from {session.provider} OAuth</p>
        </div>

        <div className="space-y-8">
          {/* Profile Picture */}
          <div className="flex items-start gap-6">
            <Avatar className="ring-foreground/10 h-24 w-24 ring-2">
              <AvatarImage src={session.user.image} alt={session.user.name} />
              <AvatarFallback className="text-2xl">{getInitials(session.user.name)}</AvatarFallback>
            </Avatar>
            <div className="flex-1 space-y-1 pt-2">
              <label className="text-foreground text-base font-semibold">Profile Picture</label>
              <p className="text-muted-foreground text-base">
                Synced from your {session.provider} account
              </p>
            </div>
          </div>

          <Separator />

          {/* Name */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <User className="text-muted-foreground h-4 w-4" />
              <label className="text-foreground text-base font-semibold">Name</label>
            </div>
            <div className="border-foreground/10 bg-muted/30 border px-4 py-3">
              <p className="text-base font-medium">{session.user.name}</p>
            </div>
            <p className="text-muted-foreground text-base">
              Synced from your {session.provider} account
            </p>
          </div>

          <Separator />

          {/* Email */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Mail className="text-muted-foreground h-4 w-4" />
              <label className="text-foreground text-base font-semibold">Email Address</label>
            </div>
            <div className="border-foreground/10 bg-muted/30 border px-4 py-3">
              <p className="text-base font-medium">{session.user.email}</p>
            </div>
            <p className="text-muted-foreground text-base">
              Primary email from your {session.provider} account
            </p>
          </div>
        </div>
      </div>

      {/* Authentication */}
      <div className="border border-foreground/10 p-8">
        <div className="mb-6">
          <div className="flex items-center gap-2">
            <Shield className="text-muted-foreground h-5 w-5" />
            <h2 className="text-xl font-semibold">Authentication</h2>
          </div>
          <p className="text-muted-foreground mt-1 text-base">Your authentication provider information</p>
        </div>

        <div className="space-y-3">
          <label className="text-foreground text-base font-semibold">OAuth Provider</label>
          <div className="border-foreground/10 bg-muted/30 flex items-center gap-3 border px-4 py-3">
            <Badge variant="outline" className="capitalize">
              {session.provider}
            </Badge>
            <span className="text-muted-foreground text-base">
              You&apos;re signed in with {session.provider}
            </span>
          </div>
        </div>
      </div>

      {/* Info Note */}
      <div className="bg-muted/50 flex items-start gap-2 border border-foreground/10 p-4">
        <Info className="text-muted-foreground mt-0.5 h-4 w-4 flex-shrink-0" />
        <p className="text-muted-foreground text-base">
          Your profile information is managed by your OAuth provider ({session.provider}). To update
          your name, email, or profile picture, please update them in your {session.provider}{' '}
          account settings.
        </p>
      </div>
    </div>
  )
}
