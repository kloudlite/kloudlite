'use client'

import { ReactNode } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { ArrowLeft } from 'lucide-react'

interface InstallationLayoutProps {
  children: ReactNode
}

export default function InstallationLayout({ children }: InstallationLayoutProps) {
  const router = useRouter()

  return (
    <div className="bg-background min-h-screen p-8">
      <div className="mx-auto w-full max-w-3xl">
        {/* Back button */}
        <div className="mb-6">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => router.push('/installations')}
            className="gap-2"
          >
            <ArrowLeft className="size-4" />
            Back to Installations
          </Button>
        </div>

        {/* Logo */}
        <div className="mb-8 flex items-center justify-center">
          <KloudliteLogo />
        </div>

        {/* Content */}
        {children}
      </div>
    </div>
  )
}
