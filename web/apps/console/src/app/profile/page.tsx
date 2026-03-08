import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getSelectedOrg } from '@/lib/console/get-selected-org'
import { getUserOrganizations } from '@/lib/console/storage'
import { InstallationsHeader } from '@/components/installations-header'
import { Avatar, AvatarFallback, AvatarImage, Badge, ScrollArea } from '@kloudlite/ui'
import { User, Mail, Shield, Info, ArrowLeft } from 'lucide-react'
import Link from 'next/link'

export default async function ProfilePage() {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const currentOrg = await getSelectedOrg(session.user.id, session.user.name, session.user.email)
  const orgs = await getUserOrganizations(session.user.id)

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  return (
    <div className="bg-background flex h-screen flex-col">
      <InstallationsHeader
        user={session.user}
        orgs={orgs.map((o) => ({ id: o.id, name: o.name, slug: o.slug }))}
        currentOrgId={currentOrg?.id}
      />

      <ScrollArea className="flex-1">
        <main className="mx-auto max-w-7xl px-6 lg:px-12 py-10">
          {/* Back */}
          <div className="mb-8">
            <Link
              href="/installations"
              className="group inline-flex items-center gap-2 text-muted-foreground hover:text-primary transition-colors duration-300 text-sm"
            >
              <ArrowLeft className="h-4 w-4 transition-transform duration-300 group-hover:-translate-x-1" />
              <span className="relative">
                Back to Installations
                <span className="absolute -bottom-0.5 left-0 right-0 h-0.5 bg-primary scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" />
              </span>
            </Link>
          </div>

          {/* Title */}
          <div className="mb-8">
            <h1 className="text-4xl lg:text-5xl font-bold tracking-tight text-foreground leading-[1.1]">Profile</h1>
            <p className="text-muted-foreground mt-2 text-[1.0625rem] leading-relaxed">
              Your personal account information
            </p>
          </div>

          {/* Profile Content */}
          <div className="space-y-6 max-w-2xl">
            {/* Profile Information Section */}
            <div>
              <div className="mb-5">
                <h2 className="text-xl font-semibold">Profile Information</h2>
                <p className="text-muted-foreground mt-1 text-base">Your account details from {session.provider} OAuth</p>
              </div>

              <div className="space-y-5">
                {/* Profile Picture */}
                <div className="flex items-start gap-6">
                  <Avatar className="ring-foreground/10 h-24 w-24 ring-1">
                    <AvatarImage src={session.user.image} alt={session.user.name} />
                    <AvatarFallback className="text-2xl">{getInitials(session.user.name)}</AvatarFallback>
                  </Avatar>
                  <div className="flex-1 space-y-1 pt-2">
                    <label className="text-foreground text-base font-medium">Profile Picture</label>
                    <p className="text-muted-foreground text-base">
                      Synced from your {session.provider} account
                    </p>
                  </div>
                </div>

                <div className="h-px bg-foreground/10" />

                {/* Name */}
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <User className="text-muted-foreground h-4 w-4" />
                    <label className="text-foreground text-base font-medium">Name</label>
                  </div>
                  <div className="border-foreground/10 bg-muted/30 border px-4 py-3">
                    <p className="text-base">{session.user.name}</p>
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
                    <label className="text-foreground text-base font-medium">Email Address</label>
                  </div>
                  <div className="border-foreground/10 bg-muted/30 border px-4 py-3">
                    <p className="text-base">{session.user.email}</p>
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
                    <label className="text-foreground text-base font-medium">Authentication Provider</label>
                  </div>
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
            </div>

            {/* Info Note */}
            <div className="bg-muted/30 flex items-start gap-3 border border-foreground/10 p-4">
              <Info className="text-muted-foreground mt-0.5 h-4 w-4 flex-shrink-0" />
              <p className="text-muted-foreground text-sm">
                Your profile information is managed by your OAuth provider ({session.provider}). To update
                your name, email, or profile picture, please update them in your {session.provider}{' '}
                account settings.
              </p>
            </div>
          </div>
        </main>
      </ScrollArea>
    </div>
  )
}
