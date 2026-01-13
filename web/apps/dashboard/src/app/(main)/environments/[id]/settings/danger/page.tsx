import { DangerSettings } from '../../../_components/danger-settings'
import { getEnvironmentDetails } from '@/lib/services/dashboard.service'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function DangerSettingsPage({ params }: PageProps) {
  const { id } = await params

  try {
    const data = await getEnvironmentDetails(id)
    const env = data.environment

    return (
      <DangerSettings
        environmentId={id}
        environmentName={`${env.spec.ownedBy}/${env.metadata.name}`}
      />
    )
  } catch (error) {
    console.error('Failed to fetch environment:', error)
    return (
      <DangerSettings
        environmentId={id}
        environmentName={id}
      />
    )
  }
}
