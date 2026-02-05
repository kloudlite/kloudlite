import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { ServicesList } from '../../_components/services-list'
import { getEnvironmentData } from '../environment-data'
import type { CompositionSpec, CompositionStatus } from '@kloudlite/types'
import type { PageProps } from '@/types/shared'

export default async function ServicesPage({ params }: PageProps) {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  // id is now the environment hash
  const { id: hash } = await params

  // Fetch environment details using server action
  const result = await getEnvironmentData(hash)

  if (result.success && result.data) {
    const data = result.data
    const environmentName = data.environment.metadata?.name || ''

    return (
      <ServicesList
        services={data.services}
        namespace={data.namespace}
        environmentName={environmentName}
        compose={data.compose as CompositionSpec | null}
        composeStatus={data.composeStatus as CompositionStatus | null}
        envHash={data.envHash || ''}
        subdomain={data.subdomain || ''}
        isEnvActive={data.isActive}
      />
    )
  }

  // Fallback if fetching fails
  return (
    <ServicesList
      services={[]}
      namespace=""
      environmentName=""
      compose={null}
      composeStatus={null}
      envHash=""
      subdomain=""
      isEnvActive={false}
    />
  )
}
