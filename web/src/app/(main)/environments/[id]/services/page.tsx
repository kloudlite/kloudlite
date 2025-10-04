import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { ServicesList } from '../../_components/services-list'
import { serviceService } from '@/lib/services/service.service'
import { environmentService } from '@/lib/services/environment.service'

interface PageProps {
  params: {
    id: string
  }
}

export default async function ServicesPage({ params }: PageProps) {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  const environmentName = params.id

  // Fetch the environment to get its target namespace
  let namespace = ''
  try {
    const environment = await environmentService.getEnvironment(environmentName)
    namespace = environment.spec.targetNamespace
  } catch (error) {
    console.error('Failed to fetch environment:', error)
    return <ServicesList services={[]} namespace={environmentName} />
  }

  // Fetch services from API using the target namespace
  let services = []
  try {
    const response = await serviceService.listServices(namespace)
    services = response.services || []
  } catch (error) {
    console.error('Failed to fetch services:', error)
    services = []
  }

  return <ServicesList services={services} namespace={namespace} />
}
