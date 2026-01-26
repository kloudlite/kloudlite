'use client'

import { ReactNode } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { GridContainer } from '@/components/grid-container'
import { ArrowLeft } from 'lucide-react'

interface InstallationLayoutProps {
  children: ReactNode
}

export default function InstallationLayout({ children }: InstallationLayoutProps) {
  const router = useRouter()

  return (
    <div className="bg-background min-h-screen px-6 py-16">
      <div className="mx-auto w-full max-w-4xl">
        <GridContainer className="border-t">
          {/* Back button and Logo */}
          <div className="border-b px-8 py-10">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.push('/installations')}
              className="mb-8 gap-2"
            >
              <ArrowLeft className="size-4" />
              Back to Installations
            </Button>

            <div className="flex items-center justify-center">
              <KloudliteLogo />
            </div>
          </div>

          {/* Content */}
          <div className="px-8 py-10">
            {children}
          </div>
        </GridContainer>
      </div>
    </div>
  )
}
