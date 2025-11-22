'use client'

import { useRouter } from 'next/navigation'
import { Key } from 'lucide-react'
import type { EnvVar } from '@kloudlite/types'
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
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium">Envvars</h3>
          <p className="text-muted-foreground text-sm">Configuration and secret envvars</p>
        </div>
        <AddEnvVarSheet environmentId={environmentId} onSuccess={() => router.refresh()} />
      </div>

      {envVars.length === 0 ? (
        <div className="bg-muted/50 rounded-lg border py-12 text-center">
          <Key className="text-muted-foreground mx-auto mb-4 h-12 w-12" />
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
