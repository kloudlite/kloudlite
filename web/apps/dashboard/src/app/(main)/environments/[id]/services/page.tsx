import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { ServicesList } from '../../_components/services-list'
import { getEnvironmentDetails } from '@/app/actions/environment.actions'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function ServicesPage({ params }: PageProps) {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  const { id } = await params

  // Fetch environment details using server action
  const result = await getEnvironmentDetails(id)

  if (result.success && result.data) {
    const data = result.data

    return (
      <ServicesList
        services={data.services}
        namespace={data.namespace}
        environmentName={id}
        compose={data.compose}
        composeStatus={data.composeStatus}
        envHash=""
        subdomain=""
        isEnvActive={data.isActive}
      />
    )
  }

  // Fallback if fetching fails
  return (
    <ServicesList
      services={[]}
      namespace={id}
      environmentName={id}
      compose={null}
      composeStatus={null}
      envHash=""
      subdomain=""
      isEnvActive={false}
    />
  )
}
