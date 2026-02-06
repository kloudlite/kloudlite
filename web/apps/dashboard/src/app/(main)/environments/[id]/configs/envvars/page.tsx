import { EnvVarsList } from '../../../_components/envvars-list'
import { getEnvVars } from '@/app/actions/environment-config'
import { getEnvironmentData } from '../../environment-data'
import { AlertCircle } from 'lucide-react'
import { Alert, AlertTitle, AlertDescription } from '@kloudlite/ui'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

// Error component
function EnvVarsError({ error }: { error: string }) {
  return (
    <Alert variant="destructive">
      <AlertCircle className="h-5 w-5" />
      <AlertTitle>Error loading envvars</AlertTitle>
      <AlertDescription>{error}</AlertDescription>
    </Alert>
  )
}

export default async function EnvVarsPage({ params }: PageProps) {
  // id is now the environment hash
  const { id: hash } = await params
  try {
    // First get the environment name from the hash
    const envResult = await getEnvironmentData(hash)
    if (!envResult.success || !envResult.data) {
      return <EnvVarsError error="Environment not found" />
    }
    const environmentName = envResult.data.environment.metadata?.name || ''

    const result = await getEnvVars(environmentName)
    const envVars = result.envVars || []

    return <EnvVarsList environmentId={environmentName} envVars={envVars} />
  } catch (error) {
    return (
      <EnvVarsError
        error={error instanceof Error ? error.message : 'Failed to load environment variables'}
      />
    )
  }
}
