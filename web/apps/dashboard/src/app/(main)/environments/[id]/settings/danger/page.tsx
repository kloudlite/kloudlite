import { DangerSettings } from '../../../_components/danger-settings'
import { getEnvironmentByHash } from '@/app/actions/environment.actions'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function DangerSettingsPage({ params }: PageProps) {
  // id is now the environment hash
  const { id: hash } = await params

  const result = await getEnvironmentByHash(hash)

  if (result.success && result.data) {
    const env = result.data.environment
    const environmentName = env.metadata?.name || ''

    return (
      <DangerSettings
        environmentId={environmentName}
        environmentName={`${env.spec?.ownedBy}/${environmentName}`}
      />
    )
  }

  // Fallback if fetching fails
  return (
    <DangerSettings
      environmentId=""
      environmentName=""
    />
  )
}
