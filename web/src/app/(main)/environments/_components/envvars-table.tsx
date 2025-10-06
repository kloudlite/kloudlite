import { Key } from 'lucide-react'
import type { EnvVar } from '@/types/environment'
import { EnvVarActions } from './envvar-actions'

interface EnvVarsTableProps {
  envVars: EnvVar[]
  environmentId: string
}

export function EnvVarsTable({ envVars, environmentId }: EnvVarsTableProps) {
  return (
    <div className="bg-card rounded-lg border overflow-hidden">
      <table className="min-w-full">
        <thead className="bg-muted/50 border-b">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Key</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Value</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Type</th>
            <th className="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody className="divide-y">
          {envVars.map((envVar) => (
            <tr key={envVar.key} className="hover:bg-muted/50">
              <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div className="flex items-center gap-2">
                  <Key className="h-4 w-4 text-muted-foreground" />
                  {envVar.key}
                </div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm font-mono">
                {envVar.type === 'secret' ? (
                  <span className="text-muted-foreground">••••••••</span>
                ) : (
                  <span className="max-w-xs truncate block">{envVar.value}</span>
                )}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm">
                {envVar.type === 'config' ? (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400">
                    Config
                  </span>
                ) : (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400">
                    Secret
                  </span>
                )}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-right text-sm space-x-1">
                <EnvVarActions
                  envVar={envVar}
                  environmentId={environmentId}
                />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
