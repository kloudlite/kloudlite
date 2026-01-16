'use client'

import { useState, useEffect, useRef } from 'react'
import Link from 'next/link'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { GetStartedButton } from '@/components/get-started-button'
import { cn } from '@kloudlite/lib'
import { Menu, X } from 'lucide-react'

interface WebsiteHeaderProps {
  currentPage?: 'home' | 'docs' | 'pricing' | 'about' | 'contact' | 'changelog' | 'workspaces' | 'environments' | 'workmachines' | 'why-kloudlite' | 'blog'
  alwaysShowBorder?: boolean
}

const navLinks = [
  { label: 'Why Kloudlite?', href: '/why-kloudlite', key: 'why-kloudlite' },
  { label: 'Docs', href: '/docs', key: 'docs' },
  { label: 'Blog', href: '/blog', key: 'blog' },
  { label: 'Pricing', href: '/pricing', key: 'pricing' },
]

export function WebsiteHeader({ currentPage, alwaysShowBorder = false }: WebsiteHeaderProps) {
  const [isScrolled, setIsScrolled] = useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const sentinelRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (alwaysShowBorder) return

    const sentinel = sentinelRef.current
    if (!sentinel) return

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
      {!alwaysShowBorder && <div ref={sentinelRef} className="h-0 w-full" aria-hidden="true" />}
      <header
        className={cn(
          'sticky top-0 z-50 bg-background/95 backdrop-blur-sm border-b transition-colors duration-200',
          showBorder ? 'border-foreground/10' : 'border-transparent'
        )}
      >
        <nav className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
          <KloudliteLogo showText={true} linkToHome={true} />

          <div className="hidden items-center gap-8 md:flex">
            {navLinks.map((link) => (
              <Link
                key={link.key}
                href={link.href}
                className={cn(
                  'text-sm transition-colors',
                  currentPage === link.key
                    ? 'text-foreground'
                    : 'text-foreground/50 hover:text-foreground'
                )}
              >
                {link.label}
              </Link>
            ))}
          </div>

          <div className="flex items-center gap-4">
            <Link
              href="https://github.com/kloudlite/kloudlite"
              target="_blank"
              rel="noopener noreferrer"
              className="hidden md:block text-foreground/50 hover:text-foreground transition-colors"
              aria-label="GitHub"
            >
              <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
                <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.17 6.839 9.49.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.604-3.369-1.34-3.369-1.34-.454-1.156-1.11-1.464-1.11-1.464-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.167 22 16.418 22 12c0-5.523-4.477-10-10-10z" />
              </svg>
            </Link>
            <GetStartedButton size="sm" className="hidden md:flex" />
            <button
              className="md:hidden p-2 -mr-2 text-foreground/60 hover:text-foreground"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              aria-label="Toggle menu"
            >
              {mobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
            </button>
          </div>
        </nav>

        {/* Mobile menu */}
        {mobileMenuOpen && (
          <div className="md:hidden border-t border-foreground/10">
            <div className="px-6 py-4 space-y-1">
              {navLinks.map((link) => (
                <Link
                  key={link.key}
                  href={link.href}
                  className={cn(
                    'block py-2.5 text-sm transition-colors',
                    currentPage === link.key
                      ? 'text-foreground'
                      : 'text-foreground/50 hover:text-foreground'
                  )}
                  onClick={() => setMobileMenuOpen(false)}
                >
                  {link.label}
                </Link>
              ))}
              <Link
                href="https://github.com/kloudlite/kloudlite"
                target="_blank"
                rel="noopener noreferrer"
                className="block py-2.5 text-foreground/50 hover:text-foreground text-sm transition-colors"
                onClick={() => setMobileMenuOpen(false)}
              >
                GitHub
              </Link>
              <div className="pt-3">
                <GetStartedButton size="sm" className="w-full" />
              </div>
            </div>
          </div>
        )}
      </header>
    </>
  )
}
