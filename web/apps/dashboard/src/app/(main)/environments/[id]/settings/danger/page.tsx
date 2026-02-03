import { DangerSettings } from '../../../_components/danger-settings'
import { getEnvironmentDetails } from '@/app/actions/environment.actions'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function DangerSettingsPage({ params }: PageProps) {
  const { id } = await params

  const result = await getEnvironmentDetails(id)

  if (result.success && result.data) {
    const env = result.data.environment

    return (
      <DangerSettings
        environmentId={id}
        environmentName={`${env.spec?.ownedBy}/${env.metadata?.name}`}
      />
    )
  }

  // Fallback if fetching fails
  return (
    <DangerSettings
      environmentId={id}
      environmentName={id}
    />
  )
}
