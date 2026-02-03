import { AccessSettings } from '../../../_components/access-settings'
import { getEnvironmentDetails } from '@/app/actions/environment.actions'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function AccessSettingsPage({ params }: PageProps) {
  const { id } = await params

  const result = await getEnvironmentDetails(id)

  if (result.success && result.data) {
    const env = result.data.environment

    return (
      <AccessSettings
        environmentId={id}
        visibility={env.spec?.visibility || 'private'}
        sharedWith={env.spec?.sharedWith || []}
        owner={env.spec?.ownedBy || 'unknown'}
      />
    )
  }

  // Fallback if fetching fails
  return (
    <AccessSettings
      environmentId={id}
      visibility="private"
      sharedWith={[]}
      owner="unknown"
    />
  )
}
