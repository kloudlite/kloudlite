import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { ServicesList } from '../../_components/services-list'
import { getEnvironmentDetails } from '@/lib/services/dashboard.service'

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

  // Single API call to get environment, services, and composition
  try {
    const data = await getEnvironmentDetails(id)

    return (
      <ServicesList
        services={data.services}
        namespace={data.namespace}
        composition={data.composition}
        envHash={data.envHash}
        subdomain={data.subdomain}
        isEnvActive={data.isActive}
      />
    )
  } catch (error) {
    console.error('Failed to fetch environment details:', error)
    return (
      <ServicesList
        services={[]}
        namespace={id}
        composition={null}
        envHash=""
        subdomain=""
        isEnvActive={false}
      />
    )
  }
}
