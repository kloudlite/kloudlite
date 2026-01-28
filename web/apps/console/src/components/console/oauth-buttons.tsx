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
            className="w-full gap-3 text-base font-medium hover:bg-foreground/[0.03] hover:border-foreground/20 transition-all duration-200 hover:shadow-sm group"
          >
            <Icon className="h-5 w-5 transition-transform group-hover:scale-110" />
            <span className="flex-1 text-left">Continue with {provider.name}</span>
            <svg
              className="h-4 w-4 text-muted-foreground group-hover:text-foreground group-hover:translate-x-0.5 transition-all opacity-0 group-hover:opacity-100"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Button>
        )
      })}
    </div>
  )
}
