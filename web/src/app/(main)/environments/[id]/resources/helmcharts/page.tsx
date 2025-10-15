import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { HelmChartsList } from '../../../_components/helmcharts-list'
import { helmChartService } from '@/lib/services/helmchart.service'
import { environmentService } from '@/lib/services/environment.service'

interface PageProps {
  params: {
    id: string
  }
}

export default async function HelmChartsPage({ params }: PageProps) {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  const { id } = await params
  const currentUser = session.user?.email || 'test-user'
  const environmentName = id

  // First, fetch the environment to get its target namespace
  let namespace = ''
  try {
    const environment = await environmentService.getEnvironment(environmentName)
    namespace = environment.spec.targetNamespace
  } catch (error) {
    console.error('Failed to fetch environment:', error)
    // If we can't get the environment, we can't fetch helm charts
    return <HelmChartsList helmCharts={[]} namespace={environmentName} user={currentUser} />
  }

  // Fetch helm charts from API using the target namespace
  let helmCharts = []
  try {
    const response = await helmChartService.listHelmCharts(namespace)
    helmCharts = response.helmCharts || []
  } catch (error) {
    console.error('Failed to fetch helm charts:', error)
    helmCharts = []
  }

  return <HelmChartsList helmCharts={helmCharts} namespace={namespace} user={currentUser} />
}
