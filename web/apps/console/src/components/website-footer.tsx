import Link from 'next/link'
import { ThemeSwitcherServer } from '@/components/theme-switcher-server'

export function WebsiteFooter() {
  return (
    <footer className="bg-background/95 supports-[backdrop-filter]:bg-background/60 border-t backdrop-blur">
      <div className="mx-auto max-w-[90rem] px-4 py-6 sm:px-6 lg:px-8">
        <div className="flex flex-wrap items-center justify-center gap-4 sm:gap-6">
          <Link
            href="/docs"
            className="text-muted-foreground hover:text-foreground text-xs sm:text-sm transition-colors"
          >
            Docs
          </Link>
          <Link
            href="/pricing"
            className="text-muted-foreground hover:text-foreground text-xs sm:text-sm transition-colors"
          >
            Pricing
          </Link>
          <Link
            href="/contact"
            className="text-muted-foreground hover:text-foreground text-xs sm:text-sm transition-colors"
          >
            Contact
          </Link>
          <Link
            href="https://github.com/kloudlite/kloudlite"
            className="text-muted-foreground hover:text-foreground text-xs sm:text-sm transition-colors"
          >
            GitHub
          </Link>
          <ThemeSwitcherServer />
        </div>
      </div>
    </footer>
  )
}
