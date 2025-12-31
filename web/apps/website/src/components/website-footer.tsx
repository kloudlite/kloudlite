'use client'

import Link from 'next/link'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { ThemeSwitcher } from '@/components/theme-switcher'

const footerLinks = {
  Product: [
    { label: 'Why Kloudlite?', href: '/why-kloudlite' },
    { label: 'Workspaces', href: '/workspaces' },
    { label: 'Environments', href: '/environments' },
    { label: 'Workmachines', href: '/workmachines' },
  ],
  Resources: [
    { label: 'Documentation', href: '/docs' },
    { label: 'Pricing', href: '/pricing' },
    { label: 'Changelog', href: '/changelog' },
  ],
  Company: [
    { label: 'About', href: '/about' },
    { label: 'Contact', href: '/contact' },
    { label: 'Branding', href: '/branding' },
    { label: 'Privacy Policy', href: '/privacy' },
    { label: 'Terms of Service', href: '/terms' },
  ],
  Connect: [
    { label: 'GitHub', href: 'https://github.com/kloudlite/kloudlite', external: true },
    { label: 'Twitter', href: 'https://twitter.com/kloudlite', external: true },
    { label: 'Discord', href: 'https://discord.gg/kloudlite', external: true },
  ],
}

export function WebsiteFooter() {
  return (
    <footer className="border-t border-foreground/10 mt-20">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-5">
          {/* Logo and tagline */}
          <div className="col-span-2 md:col-span-1">
            <KloudliteLogo showText={true} linkToHome={true} />
            <p className="mt-4 text-sm text-foreground/40 leading-relaxed">
              Cloud development environments designed to reduce the development loop.
            </p>
          </div>

          {/* Link columns */}
          {Object.entries(footerLinks).map(([category, links]) => (
            <div key={category}>
              <h3 className="text-foreground text-sm font-medium">{category}</h3>
              <ul className="mt-4 space-y-3">
                {links.map((link) => (
                  <li key={link.label}>
                    <Link
                      href={link.href}
                      className="text-foreground/40 hover:text-foreground text-sm transition-colors"
                      {...('external' in link && link.external ? { target: '_blank', rel: 'noopener noreferrer' } : {})}
                    >
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Bottom bar */}
        <div className="mt-16 pt-8 border-t border-foreground/10 flex flex-col sm:flex-row items-center justify-between gap-4">
          <p className="text-foreground/30 text-sm">
            &copy; {new Date().getFullYear()} Kloudlite. All rights reserved.
          </p>
          <ThemeSwitcher />
        </div>
      </div>
    </footer>
  )
}
