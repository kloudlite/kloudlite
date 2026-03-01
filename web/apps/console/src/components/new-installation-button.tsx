'use client'

import { useRouter } from 'next/navigation'
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@kloudlite/ui'
import { Plus, Cloud, Server, CreditCard } from 'lucide-react'

interface NewInstallationButtonProps {
  hasActiveSubscription: boolean
}

export function NewInstallationButton({ hasActiveSubscription }: NewInstallationButtonProps) {
  const router = useRouter()

  if (!hasActiveSubscription) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              size="default"
              variant="outline"
              className="gap-2"
              onClick={() => router.push('/installations/settings/billing')}
            >
              <CreditCard className="h-4 w-4" />
              Subscribe to Create
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Subscribe to a plan to create installations</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }

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
