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
    colorClass: 'bg-github text-github-foreground hover:bg-github/90 border-github',
  },
  {
    id: 'google',
    name: 'Google',
    icon: SiGoogle,
    colorClass: 'bg-google text-google-foreground hover:bg-google/90 border-google',
  },
  {
    id: 'microsoft-entra-id',
    name: 'Microsoft',
    icon: MicrosoftIcon,
    colorClass: 'bg-microsoft text-microsoft-foreground hover:bg-microsoft/90 border-microsoft',
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
            asChild
            size="lg"
            className={`w-full gap-3 text-base font-medium transition-all duration-200 shadow-sm hover:shadow-md group ${provider.colorClass}`}
          >
            <a href={`/api/oauth/${provider.id}`}>
              <Icon className="h-5 w-5 transition-transform group-hover:scale-110" />
              <span className="flex-1 text-left">Continue with {provider.name}</span>
              <svg
                className="h-4 w-4 opacity-80 group-hover:opacity-100 group-hover:translate-x-0.5 transition-all"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </a>
          </Button>
        )
      })}
    </div>
  )
}
