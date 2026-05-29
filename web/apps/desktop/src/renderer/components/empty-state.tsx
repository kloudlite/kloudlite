import { type ReactNode } from 'react'

export function KloudliteLogo({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 130 131" fill="none" xmlns="http://www.w3.org/2000/svg" className={className}>
      <path d="M51.9912 66.6496C51.2636 65.9244 51.2636 64.7486 51.9912 64.0235L89.4072 26.7312C90.1348 26.006 91.3145 26.006 92.042 26.7312L129.458 64.0237C130.186 64.7489 130.186 65.9246 129.458 66.6498L92.0423 103.942C91.3147 104.667 90.135 104.667 89.4074 103.942L51.9912 66.6496Z" fill="currentColor" opacity="0.5"/>
      <path d="M66.5331 1.04291C65.8055 0.317729 64.6259 0.317729 63.8983 1.04291L0.545688 64.186C-0.181896 64.9111 -0.181896 66.0869 0.545688 66.8121L63.8983 129.955C64.6259 130.68 65.8055 130.68 66.5331 129.955L76.9755 119.547C77.7031 118.822 77.7031 117.646 76.9755 116.921L26.4574 66.5701C25.7298 65.8449 25.7298 64.6692 26.4574 63.944L76.7327 13.8349C77.4603 13.1097 77.4603 11.934 76.7327 11.2088L66.5331 1.04291Z" fill="currentColor" opacity="0.8"/>
    </svg>
  )
}

interface EmptyStateProps {
  title: string
  description?: string | ReactNode
  action?: {
    label: string
    onClick: () => void
  }
}

export function EmptyState({ title, description, action }: EmptyStateProps) {
  return (
    <div className="flex h-full items-center justify-center bg-background px-6">
      <div className="flex max-w-sm flex-col items-center">
        <KloudliteLogo className="h-14 w-14 text-primary/15" />
        <h3 className="mt-5 text-[15px] font-semibold text-foreground/50">{title}</h3>
        {description && (
          <p className="mt-1.5 text-center text-[13px] leading-relaxed text-muted-foreground/70">
            {description}
          </p>
        )}
        {action && (
          <button
            className="mt-5 rounded-lg bg-primary px-5 py-2 text-[12px] font-medium text-primary-foreground transition-colors hover:bg-primary/90"
            onClick={action.onClick}
          >
            {action.label}
          </button>
        )}
      </div>
    </div>
  )
}
