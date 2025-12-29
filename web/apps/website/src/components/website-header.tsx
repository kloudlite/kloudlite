'use client'

import { useState, useEffect, useRef } from 'react'
import Link from 'next/link'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { GetStartedButton } from '@/components/get-started-button'
import { cn } from '@kloudlite/lib'
import { Menu, X } from 'lucide-react'

interface WebsiteHeaderProps {
  currentPage?: 'home' | 'docs' | 'pricing'
  alwaysShowBorder?: boolean
  showSearch?: boolean
}

export function WebsiteHeader({ currentPage, alwaysShowBorder = false, showSearch: _showSearch = false }: WebsiteHeaderProps) {
  const [isScrolled, setIsScrolled] = useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const sentinelRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (alwaysShowBorder) return

    const sentinel = sentinelRef.current
    if (!sentinel) return

    // Find the ScrollArea viewport as the root
    const scrollAreaViewport = document.querySelector('[data-radix-scroll-area-viewport]')

    const observer = new IntersectionObserver(
      ([entry]) => {
        setIsScrolled(!entry.isIntersecting)
      },
      {
        threshold: 0,
        root: scrollAreaViewport || null
      }
    )

    observer.observe(sentinel)

    return () => {
      observer.disconnect()
    }
  }, [alwaysShowBorder])

  const showBorder = alwaysShowBorder || isScrolled

  return (
    <>
      {/* Sentinel for scroll detection - must be before sticky header */}
      {!alwaysShowBorder && <div ref={sentinelRef} className="h-0 w-full" aria-hidden="true" />}
      <header
        className={cn(
          'sticky top-0 z-50 bg-background/80 backdrop-blur-xl border-b transition-colors duration-200',
          showBorder ? 'border-foreground/10' : 'border-transparent'
        )}
      >
        <nav className="mx-auto flex h-16 max-w-[90rem] items-center justify-between px-6 lg:px-8">
        <div className="flex items-center gap-8 lg:gap-12">
          <KloudliteLogo showText={true} linkToHome={true} />
          <div className="hidden items-center gap-8 md:flex">
            <Link
              href="/docs"
              className={
                currentPage === 'docs'
                  ? 'text-foreground text-sm font-medium transition-all duration-100 active:translate-y-0.5'
                  : 'text-foreground/60 hover:text-foreground text-sm transition-all duration-100 active:translate-y-0.5'
              }
            >
              Docs
            </Link>
            <Link
              href="/pricing"
              className={
                currentPage === 'pricing'
                  ? 'text-foreground text-sm font-medium transition-all duration-100 active:translate-y-0.5'
                  : 'text-foreground/60 hover:text-foreground text-sm transition-all duration-100 active:translate-y-0.5'
              }
            >
              Pricing
            </Link>
          </div>
        </div>

        <div className="flex items-center gap-4">
          <GetStartedButton size="sm" className="hidden md:flex" />
          <button
            className="md:hidden p-2 text-foreground/60 hover:text-foreground"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            aria-label="Toggle menu"
          >
            {mobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </button>
        </div>
        </nav>

        {/* Mobile menu */}
        {mobileMenuOpen && (
          <div className="md:hidden bg-background/95 backdrop-blur-xl">
            <div className="px-6 py-4 space-y-4">
              <Link
                href="/docs"
                className={cn(
                  'block text-sm transition-colors',
                  currentPage === 'docs'
                    ? 'text-foreground font-medium'
                    : 'text-foreground/60 hover:text-foreground'
                )}
                onClick={() => setMobileMenuOpen(false)}
              >
                Docs
              </Link>
              <Link
                href="/pricing"
                className={cn(
                  'block text-sm transition-colors',
                  currentPage === 'pricing'
                    ? 'text-foreground font-medium'
                    : 'text-foreground/60 hover:text-foreground'
                )}
                onClick={() => setMobileMenuOpen(false)}
              >
                Pricing
              </Link>
              <div className="pt-2">
                <GetStartedButton size="sm" className="w-full" />
              </div>
            </div>
          </div>
        )}
      </header>
    </>
  )
}
