import { Link } from '@/components/ui/link'
import { ThemeToggleClient } from '@/components/theme-toggle-client'

export function Footer() {
  return (
    <footer className="border-t border-border">
      <div className="max-w-6xl mx-auto px-6 py-12">
        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-8">
          <div className="flex items-center gap-4">
            <p className="text-sm text-muted-foreground">
              © {new Date().getFullYear()} Kloudlite. All rights reserved.
            </p>
            <ThemeToggleClient />
          </div>
          <nav className="flex flex-wrap gap-6">
            <Link href="/terms" className="text-sm text-muted-foreground hover:text-primary">
              Terms
            </Link>
            <Link href="/privacy" className="text-sm text-muted-foreground hover:text-primary">
              Privacy
            </Link>
            <Link href="/docs" className="text-sm text-muted-foreground hover:text-primary">
              Documentation
            </Link>
            <Link href="https://github.com/kloudlite/kloudlite" className="text-sm text-muted-foreground hover:text-primary">
              GitHub
            </Link>
          </nav>
        </div>
      </div>
    </footer>
  )
}