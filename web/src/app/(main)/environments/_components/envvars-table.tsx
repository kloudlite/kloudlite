import { Key } from 'lucide-react'
import type { EnvVar } from '@/types/environment'
import { EnvVarActions } from './envvar-actions'

interface EnvVarsTableProps {
  envVars: EnvVar[]
  environmentId: string
}

export function EnvVarsTable({ envVars, environmentId }: EnvVarsTableProps) {
  return (
    <div className="bg-card overflow-hidden rounded-lg border">
      <table className="min-w-full">
        <thead className="bg-muted/50 border-b">
          <tr>
            <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
              Key
            </th>
            <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
              Value
            </th>
            <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
              Type
            </th>
            <th className="text-muted-foreground px-6 py-3 text-right text-xs font-medium tracking-wider uppercase">
              Actions
            </th>
          </tr>
        </thead>
        <tbody className="divide-y">
          {envVars.map((envVar) => (
            <tr key={envVar.key} className="hover:bg-muted/50">
              <td className="px-6 py-4 text-sm font-medium whitespace-nowrap">
                <div className="flex items-center gap-2">
                  <Key className="text-muted-foreground h-4 w-4" />
                  {envVar.key}
                </div>
              </td>
              <td className="px-6 py-4 font-mono text-sm whitespace-nowrap">
                {envVar.type === 'secret' ? (
                  <span className="text-muted-foreground">••••••••</span>
                ) : (
                  <span className="block max-w-xs truncate">{envVar.value}</span>
                )}
              </td>
              <td className="px-6 py-4 text-sm whitespace-nowrap">
                {envVar.type === 'config' ? (
                  <span className="inline-flex items-center rounded-full bg-info/10 px-2.5 py-0.5 text-xs font-medium text-info dark:bg-info/20">
                    Config
                  </span>
                ) : (
                  <span className="inline-flex items-center rounded-full bg-purple-100 px-2.5 py-0.5 text-xs font-medium text-purple-800 dark:bg-purple-900/30 dark:text-purple-400">
                    Secret
                  </span>
                )}
              </td>
              <td className="space-x-1 px-6 py-4 text-right text-sm whitespace-nowrap">
                <EnvVarActions envVar={envVar} environmentId={environmentId} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
