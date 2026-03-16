'use client'

import Link from 'next/link'
import { Button } from '@kloudlite/ui'
import { Plus } from 'lucide-react'

export function NewInstallationButton() {
  return (
    <Button size="default" asChild>
      <Link href="/installations/new">
        <Plus className="h-4 w-4" />
        New Installation
      </Link>
    </Button>
  )
}
