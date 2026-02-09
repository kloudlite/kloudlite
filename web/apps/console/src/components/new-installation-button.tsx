'use client'

import { useRouter } from 'next/navigation'
import { Button, DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@kloudlite/ui'
import { Plus, Cloud, Server } from 'lucide-react'

export function NewInstallationButton() {
  const router = useRouter()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button size="default">
          <Plus className="h-4 w-4" />
          New Installation
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-64">
        <DropdownMenuItem
          className="cursor-pointer py-3 px-3"
          onClick={() => router.push('/installations/new-kl-cloud')}
        >
          <Cloud className="size-4 text-primary" />
          <div>
            <p className="font-medium">Kloudlite Cloud</p>
            <p className="text-xs text-muted-foreground">We manage the infrastructure for you</p>
          </div>
        </DropdownMenuItem>
        <DropdownMenuItem
          className="cursor-pointer py-3 px-3"
          onClick={() => router.push('/installations/new-byoc')}
        >
          <Server className="size-4 text-primary" />
          <div>
            <p className="font-medium">Bring your own Cloud</p>
            <p className="text-xs text-muted-foreground">Deploy to your own AWS, GCP, or Azure</p>
          </div>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
