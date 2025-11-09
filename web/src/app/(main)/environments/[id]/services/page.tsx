import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { ServicesList } from '../../_components/services-list'
import { serviceService } from '@/lib/services/service.service'
import { environmentService } from '@/lib/services/environment.service'
import { serviceInterceptService } from '@/lib/services/serviceintercept.service'
import { compositionService } from '@/lib/services/composition.service'
import type { K8sService } from '@/types/service'
import type { ServiceIntercept } from '@/types/serviceintercept'
import type { Composition } from '@/types/composition'

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
  const environmentName = id

  // Fetch the environment to get its target namespace
  let namespace = ''
  try {
    const environment = await environmentService.getEnvironment(environmentName)
    namespace = environment.spec.targetNamespace || ''
  } catch (error) {
    console.error('Failed to fetch environment:', error)
    return (
      <div className="mx-auto max-w-7xl px-6 py-8">
        <ServicesList
          services={[]}
          namespace={environmentName}
          serviceIntercepts={[]}
          composition={null}
        />
      </div>
    )
  }

  // Fetch services from API using the target namespace
  let services: K8sService[] = []
  try {
    const response = await serviceService.listServices(namespace)
    services = response.services || []
  } catch (error) {
    console.error('Failed to fetch services:', error)
    services = []
  }

  // Fetch service intercepts from API
  let serviceIntercepts: ServiceIntercept[] = []
  try {
    const response = await serviceInterceptService.listServiceIntercepts(namespace)
    serviceIntercepts = response.serviceIntercepts || []
  } catch (error) {
    console.error('Failed to fetch service intercepts:', error)
    serviceIntercepts = []
  }

  // Fetch the main composition
  let composition: Composition | null = null
  try {
    composition = await compositionService.getComposition(namespace, 'main-composition')
  } catch {
    // Composition doesn't exist yet, that's okay
    console.log('Main composition not found, will be created on first save')
  }

  return (
    <ServicesList
      services={services}
      namespace={namespace}
      serviceIntercepts={serviceIntercepts}
      composition={composition}
    />
  )
}
