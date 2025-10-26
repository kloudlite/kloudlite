import { KloudliteLogo } from '@/components/kloudlite-logo'
import Link from 'next/link'
import { Search } from 'lucide-react'
import { DocsSidebar } from './_components/docs-sidebar'
import { getTheme } from '@/lib/theme-server'
import { GetStartedButton } from '@/components/get-started-button'

export default async function DocsLayout({ children }: { children: React.ReactNode }) {
  const theme = await getTheme()
  return (
    <div className="bg-background flex min-h-screen flex-col">
      {/* Top Navigation Header */}
      <header className="bg-background/95 supports-[backdrop-filter]:bg-background/60 sticky top-0 z-50 border-b backdrop-blur">
        <nav className="mx-auto flex h-16 max-w-[90rem] items-center justify-between px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-6 lg:gap-8">
            <KloudliteLogo showText={true} linkToHome={true} />
            <div className="hidden items-center gap-6 md:flex">
              <Link href="/docs" className="text-foreground text-sm font-medium">
                Docs
              </Link>
              <Link
                href="/pricing"
                className="text-muted-foreground hover:text-foreground text-sm font-medium transition-colors"
              >
                Pricing
              </Link>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <button className="bg-muted/50 text-muted-foreground hover:bg-muted hidden items-center gap-2 rounded-lg border px-3 py-1.5 text-sm transition-colors md:flex">
              <Search className="h-4 w-4" />
              <span>Search...</span>
              <kbd className="bg-background ml-auto hidden rounded px-1.5 text-xs font-medium lg:inline-block">
                ⌘K
              </kbd>
            </button>
            <GetStartedButton size="sm" className="hidden sm:flex" />
          </div>
        </nav>
      </header>

      {/* Main Content Area with Sidebar */}
      <div className="mx-auto w-full max-w-[90rem]">
        <div className="flex">
          {/* Sidebar */}
          <DocsSidebar initialTheme={theme} />

          {/* Main Content */}
          <div className="flex-1">{children}</div>
        </div>
      </div>
    </div>
  )
}
