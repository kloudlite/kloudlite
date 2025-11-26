import { EnvVarsList } from '../../../_components/envvars-list'
import { getEnvVars } from '@/app/actions/environment-config'
import { AlertCircle } from 'lucide-react'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

// Error component
function EnvVarsError({ error }: { error: string }) {
  return (
    <div className="rounded-lg border border-red-200 bg-red-50 p-4">
      <div className="flex items-center gap-2 text-red-800">
        <AlertCircle className="h-5 w-5" />
        <span className="font-medium">Error loading envvars</span>
      </div>
      <p className="mt-2 text-sm text-red-700">{error}</p>
    </div>
  )
}

export default async function EnvVarsPage({ params }: PageProps) {
  const { id } = await params
  try {
    const result = await getEnvVars(id)
    const envVars = result.envVars || []

    return <EnvVarsList environmentId={id} envVars={envVars} />
  } catch (error) {
    return (
      <EnvVarsError
        error={error instanceof Error ? error.message : 'Failed to load environment variables'}
      />
    )
  }
}
