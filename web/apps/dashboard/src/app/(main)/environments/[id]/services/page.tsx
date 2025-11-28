import { redirect } from 'next/navigation'
import { headers } from 'next/headers'
import { getSession } from '@/lib/get-session'
import { ServicesList } from '../../_components/services-list'
import { serviceService } from '@/lib/services/service.service'
import { environmentService } from '@/lib/services/environment.service'
import { compositionService } from '@/lib/services/composition.service'
import type { K8sService } from '@kloudlite/types'
import type { Composition } from '@kloudlite/types'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

// Parse subdomain and domain from hostname
function parseDomainInfo(hostname: string): { subdomain: string; domain: string } {
  const baseDomain = 'khost.dev'
  const hostParts = hostname.split('.')
  const baseParts = baseDomain.split('.')

  let subdomain = ''
  if (hostParts.length > baseParts.length) {
    subdomain = hostParts[hostParts.length - baseParts.length - 1]
  }

  return { subdomain, domain: baseDomain }
}

export default async function ServicesPage({ params }: PageProps) {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  const { id } = await params
  const environmentName = id

  // Get hostname from headers for domain parsing
  const headersList = await headers()
  const host = headersList.get('host') || ''
  const { subdomain, domain } = parseDomainInfo(host)

  // Fetch the environment to get its target namespace and owner
  let namespace = ''
  let owner = ''
  try {
    const environment = await environmentService.getEnvironment(environmentName)
    namespace = environment.spec.targetNamespace || ''
    owner = environment.spec.ownedBy || ''
  } catch (error) {
    console.error('Failed to fetch environment:', error)
    return (
      <div className="mx-auto max-w-7xl px-6 py-8">
        <ServicesList
          services={[]}
          namespace={environmentName}
          composition={null}
          environmentName={environmentName}
          owner=""
          subdomain={subdomain}
          domain={domain}
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

  // Fetch the main composition (service intercepts are part of composition)
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
      composition={composition}
      environmentName={environmentName}
      owner={owner}
      subdomain={subdomain}
      domain={domain}
    />
  )
}
