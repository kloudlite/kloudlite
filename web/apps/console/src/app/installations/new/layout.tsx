'use client'

import { ReactNode } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { ArrowLeft } from 'lucide-react'

interface InstallationLayoutProps {
  children: ReactNode
}

export default function InstallationLayout({ children }: InstallationLayoutProps) {
  const router = useRouter()

  return (
    <div className="bg-background min-h-screen px-6 lg:px-12 py-16">
      <div className="mx-auto w-full max-w-4xl">
        <div className="border border-border shadow-sm">
          {/* Back button */}
          <div className="border-b border-border px-6 lg:px-12 py-6 bg-muted/20">
            <Button
              variant="ghost"
              onClick={() => router.push('/installations')}
              className="group -ml-3 text-muted-foreground hover:text-primary transition-colors duration-300 text-sm h-auto p-3"
            >
              <ArrowLeft className="h-4 w-4 transition-transform duration-300 group-hover:-translate-x-1" />
              <span className="relative">
                Back to Installations
                <span className="absolute -bottom-0.5 left-0 right-0 h-0.5 bg-primary scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" />
              </span>
            </Button>
          </div>

          {/* Content */}
          <div className="px-6 lg:px-12 py-12 bg-background">
            {children}
          </div>
        </div>
      </div>
    </div>
  )
}
