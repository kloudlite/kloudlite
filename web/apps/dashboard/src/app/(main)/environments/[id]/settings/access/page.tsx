import { AccessSettings } from '../../../_components/access-settings'
import { getEnvironmentData } from '../../environment-data'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function AccessSettingsPage({ params }: PageProps) {
  // id is now the environment hash
  const { id: hash } = await params

  const result = await getEnvironmentData(hash)

  if (result.success && result.data) {
    const env = result.data.environment
    const environmentName = env.metadata?.name || ''

    return (
      <AccessSettings
        environmentId={environmentName}
        visibility={env.spec?.visibility || 'private'}
        sharedWith={env.spec?.sharedWith || []}
        owner={env.spec?.ownedBy || 'unknown'}
      />
    )
  }

  // Fallback if fetching fails
  return (
    <AccessSettings
      environmentId=""
      visibility="private"
      sharedWith={[]}
      owner="unknown"
    />
  )
}
