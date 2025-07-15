import { Link } from '@/components/ui/link'
import { Button } from '@/components/ui/button'

export function Navbar() {
  return (
    <header className="sticky top-0 z-50 w-full bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 border-b border-border">
      <nav className="max-w-6xl mx-auto flex h-16 items-center justify-between px-6">
        <Link href="/" className="text-lg font-bold">
          Kloudlite
        </Link>
        <div className="flex items-center gap-6">
          <Link href="/docs" className="text-sm hover:text-primary transition-colors">
            Docs
          </Link>
          <Link href="/auth/login" className="text-sm hover:text-primary transition-colors">
            Sign in
          </Link>
          <Button asChild>
            <Link href="/auth/signup">
              Get started
            </Link>
          </Button>
        </div>
      </nav>
    </header>
  )
}