'use client'

import Link from 'next/link'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { ThemeSwitcher } from '@/components/theme-switcher'

const footerLinks = {
  Product: [
    { label: 'Workspaces', href: '/docs/workspaces' },
    { label: 'Environments', href: '/docs/environments' },
    { label: 'Packages', href: '/docs/packages' },
    { label: 'Intercepts', href: '/docs/intercepts' },
  ],
  Resources: [
    { label: 'Documentation', href: '/docs' },
    { label: 'Pricing', href: '/pricing' },
    { label: 'Changelog', href: '/changelog' },
  ],
  Company: [
    { label: 'About', href: '/about' },
    { label: 'Contact', href: '/contact' },
  ],
  Social: [
    { label: 'GitHub', href: 'https://github.com/kloudlite/kloudlite', external: true },
    { label: 'Twitter', href: 'https://twitter.com/kloudlite', external: true },
    { label: 'Discord', href: 'https://discord.gg/kloudlite', external: true },
  ],
}

export function WebsiteFooter() {
  return (
    <footer className="border-t border-foreground/[0.06]">
      <div className="mx-auto max-w-[90rem] px-6 py-16 lg:px-8">
        <div className="grid grid-cols-2 gap-8 sm:grid-cols-4 lg:grid-cols-5">
          {/* Logo column */}
          <div className="col-span-2 sm:col-span-4 lg:col-span-1 mb-8 lg:mb-0">
            <KloudliteLogo showText={false} linkToHome={true} />
          </div>

          {/* Link columns */}
          {Object.entries(footerLinks).map(([category, links]) => (
            <div key={category}>
              <h3 className="text-foreground text-sm font-medium mb-4">{category}</h3>
              <ul className="space-y-3">
                {links.map((link) => (
                  <li key={link.label}>
                    <Link
                      href={link.href}
                      className="text-foreground/50 hover:text-foreground text-sm transition-all duration-100 active:translate-y-0.5"
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
        <div className="mt-16 pt-8 border-t border-foreground/[0.06] flex flex-col sm:flex-row items-center justify-between gap-4">
          <p className="text-foreground/40 text-sm">
            &copy; {new Date().getFullYear()} Kloudlite
          </p>
          <ThemeSwitcher />
        </div>
      </div>
    </footer>
  )
}
