'use client'

import { Button } from '@kloudlite/ui'
import { SiGithub, SiGoogle } from 'react-icons/si'
import { Building2 } from 'lucide-react'
import { signIn } from 'next-auth/react'

const providers = [
  {
    id: 'github',
    name: 'GitHub',
    icon: SiGithub,
  },
  {
    id: 'google',
    name: 'Google',
    icon: SiGoogle,
  },
  {
    id: 'microsoft-entra-id',
    name: 'Microsoft',
    icon: Building2,
  },
]

export function OAuthButtons() {
  return (
    <div className="space-y-3">
      {providers.map((provider) => {
        const Icon = provider.icon
        return (
          <Button
            key={provider.id}
            onClick={() => signIn(provider.id, { callbackUrl: '/installations' })}
            variant="outline"
            className="w-full gap-3"
          >
            <Icon className="h-5 w-5" />
            <span>Continue with {provider.name}</span>
          </Button>
        )
      })}
    </div>
  )
}
