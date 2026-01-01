import { AccessSettings } from '../../../_components/access-settings'
import { getEnvironmentDetails } from '@/lib/services/dashboard.service'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function AccessSettingsPage({ params }: PageProps) {
  const { id } = await params

  try {
    const data = await getEnvironmentDetails(id)
    const env = data.environment

    return (
      <AccessSettings
        environmentId={id}
        visibility={env.spec.visibility || 'private'}
        sharedWith={env.spec.sharedWith || []}
        owner={env.spec.ownedBy}
      />
    )
  } catch (error) {
    console.error('Failed to fetch environment:', error)
    return (
      <AccessSettings
        environmentId={id}
        visibility="private"
        sharedWith={[]}
        owner="unknown"
      />
    )
  }
}
