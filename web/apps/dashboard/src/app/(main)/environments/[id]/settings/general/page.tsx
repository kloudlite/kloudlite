import { GeneralSettings } from '../../../_components/general-settings'
import { getEnvironmentByHash } from '@/app/actions/environment.actions'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function GeneralSettingsPage({ params }: PageProps) {
  // id is now the environment hash
  const { id: hash } = await params

  const result = await getEnvironmentByHash(hash)

  if (result.success && result.data) {
    const env = result.data.environment
    const environmentName = env.metadata?.name || ''

    return (
      <GeneralSettings
        environmentId={environmentName}
        environmentName={environmentName}
        description={env.metadata?.annotations?.['kloudlite.io/description'] || ''}
        ownedBy={env.spec?.ownedBy || 'unknown'}
      />
    )
  }

  // Fallback if fetching fails
  return (
    <GeneralSettings
      environmentId=""
      environmentName=""
      description=""
      ownedBy="unknown"
    />
  )
}
