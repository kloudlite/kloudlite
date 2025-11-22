import { Button } from '@kloudlite/ui'
import Link from 'next/link'

interface GetStartedButtonProps {
  size?: 'sm' | 'lg' | 'default'
  className?: string
  variant?: 'default' | 'outline'
}

export function GetStartedButton({
  size = 'default',
  className,
  variant = 'default',
}: GetStartedButtonProps) {
  return (
    <Button asChild size={size} className={className} variant={variant}>
      <Link href="https://console.kloudlite.io" target="_blank" rel="noopener noreferrer">
        Access Console
      </Link>
    </Button>
  )
}
