import { Link } from '@/components/ui/link'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { LAYOUT } from '@/lib/constants/layout'

export function Navbar() {
  return (
    <header className="sticky top-0 z-50 w-full bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 border-b border-border">
      <nav className={cn("max-w-6xl mx-auto flex h-14 sm:h-16 items-center justify-between", LAYOUT.PADDING.HEADER_X)}>
        <Link href="/" className="text-base sm:text-lg font-bold">
          Kloudlite
        </Link>
        <div className="flex items-center gap-3 sm:gap-6">
          <Link href="/docs" className="text-sm hover:text-primary transition-colors hidden sm:inline-flex">
            Docs
          </Link>
          <Link href="/auth/login" className="text-sm hover:text-primary transition-colors hidden sm:inline-flex">
            Sign in
          </Link>
          <Button size="sm" asChild>
            <Link href="/auth/signup">
              Get started
            </Link>
          </Button>
        </div>
      </nav>
    </header>
  )
}