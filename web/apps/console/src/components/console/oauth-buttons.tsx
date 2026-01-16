'use client'

import { Button } from '@kloudlite/ui'
import { SiGithub, SiGoogle } from 'react-icons/si'

// Microsoft logo component
function MicrosoftIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 23 23" fill="currentColor">
      <path d="M0 0h11v11H0z" />
      <path d="M12 0h11v11H12z" />
      <path d="M0 12h11v11H0z" />
      <path d="M12 12h11v11H12z" />
    </svg>
  )
}

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
    icon: MicrosoftIcon,
  },
]

export function OAuthButtons() {
  const handleClick = (providerId: string) => {
    window.location.href = `/api/oauth/${providerId}`
  }

  return (
    <div className="space-y-3">
      {providers.map((provider) => {
        const Icon = provider.icon
        return (
          <Button
            key={provider.id}
            onClick={() => handleClick(provider.id)}
            variant="outline"
            size="lg"
            className="w-full gap-3 text-base"
          >
            <Icon className="h-5 w-5" />
            <span>Continue with {provider.name}</span>
          </Button>
        )
      })}
    </div>
  )
}
