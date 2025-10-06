'use client'

import { useRouter } from 'next/navigation'
import { Key } from 'lucide-react'
import type { EnvVar } from '@/types/environment'
import { AddEnvVarSheet } from './add-envvar-sheet'
import { EnvVarsTable } from './envvars-table'

interface EnvVarsListProps {
  environmentId: string
  envVars: EnvVar[]
}

export function EnvVarsList({ environmentId, envVars }: EnvVarsListProps) {
  const router = useRouter()

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">Envvars</h3>
          <p className="text-sm text-muted-foreground">Configuration and secret envvars</p>
        </div>
        <AddEnvVarSheet environmentId={environmentId} onSuccess={() => router.refresh()} />
      </div>

      {envVars.length === 0 ? (
        <div className="text-center py-12 bg-muted/50 rounded-lg border">
          <Key className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No envvars</p>
          <div className="mt-4">
            <AddEnvVarSheet environmentId={environmentId} onSuccess={() => router.refresh()} />
          </div>
        </div>
      ) : (
        <EnvVarsTable envVars={envVars} environmentId={environmentId} />
      )}
    </div>
  )
}
