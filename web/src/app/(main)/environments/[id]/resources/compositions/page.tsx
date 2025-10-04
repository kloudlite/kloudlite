import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { CompositionsList } from '@/components/compositions-list'
import { compositionService } from '@/lib/services/composition.service'
import { environmentService } from '@/lib/services/environment.service'

interface PageProps {
  params: {
    id: string
  }
}

export default async function CompositionsPage({ params }: PageProps) {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  const currentUser = session.user?.email || 'test-user'
  const environmentName = params.id

  // First, fetch the environment to get its target namespace
  let namespace = ''
  try {
    const environment = await environmentService.getEnvironment(environmentName)
    namespace = environment.spec.targetNamespace
  } catch (error) {
    console.error('Failed to fetch environment:', error)
    // If we can't get the environment, we can't fetch compositions
    return <CompositionsList compositions={[]} namespace={environmentName} user={currentUser} />
  }

  // Fetch compositions from API using the target namespace
  let compositions = []
  try {
    const response = await compositionService.listCompositions(namespace)
    compositions = response.compositions || []
  } catch (error) {
    console.error('Failed to fetch compositions:', error)
    compositions = []
  }

  return <CompositionsList compositions={compositions} namespace={namespace} user={currentUser} />
}
