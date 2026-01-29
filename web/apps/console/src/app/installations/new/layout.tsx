'use client'

import { ReactNode } from 'react'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'

interface InstallationLayoutProps {
  children: ReactNode
}

export default function InstallationLayout({ children }: InstallationLayoutProps) {
  return (
    <div className="bg-background min-h-screen">
      <div className="mx-auto max-w-7xl px-6 lg:px-8 py-8">
        {/* Back button */}
        <div className="mb-8">
          <Link
            href="/installations"
            className="group inline-flex items-center gap-2 text-muted-foreground hover:text-primary transition-colors duration-300 text-sm"
          >
            <ArrowLeft className="h-4 w-4 transition-transform duration-300 group-hover:-translate-x-1" />
            <span className="relative">
              Back to Installations
              <span className="absolute -bottom-0.5 left-0 right-0 h-0.5 bg-primary scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" />
            </span>
          </Link>
        </div>

        {/* Content */}
        {children}
      </div>
    </div>
  )
}
