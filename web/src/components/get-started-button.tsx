import { auth } from '@/lib/auth'
import { Button } from '@/components/ui/button'
import Link from 'next/link'

interface GetStartedButtonProps {
  size?: 'sm' | 'lg' | 'default'
  className?: string
  variant?: 'default' | 'outline'
}

export async function GetStartedButton({
  size = 'default',
  className,
  variant = 'default'
}: GetStartedButtonProps) {
  const session = await auth()

  // If user is authenticated, show "Go to Dashboard" instead
  if (session?.user) {
    return (
      <Button asChild size={size} className={className} variant={variant}>
        <Link href="/">Go to Dashboard</Link>
      </Button>
    )
  }

  // Otherwise show "Get Started" that goes to access-console page
  return (
    <Button asChild size={size} className={className} variant={variant}>
      <Link href="/installations/login">Get Started</Link>
    </Button>
  )
}
