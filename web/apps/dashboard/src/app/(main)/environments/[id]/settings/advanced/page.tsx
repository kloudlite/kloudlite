import { AdvancedSettings } from '../../../_components/advanced-settings'
import { getEnvironmentData } from '../../environment-data'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function AdvancedSettingsPage({ params }: PageProps) {
  // id is now the environment hash
  const { id: hash } = await params

  const result = await getEnvironmentData(hash)

  if (result.success && result.data) {
    const env = result.data.environment
    const environmentName = env.metadata?.name || ''

    return (
      <AdvancedSettings
        environmentId={environmentName}
        resourceLimits={{
          cpu: env.spec?.resourceQuotas?.['limits.cpu'] || '',
          memory: env.spec?.resourceQuotas?.['limits.memory'] || '',
        }}
      />
    )
  }

  // Fallback if fetching fails
  return (
    <AdvancedSettings
      environmentId=""
      resourceLimits={{ cpu: '', memory: '' }}
    />
  )
}
