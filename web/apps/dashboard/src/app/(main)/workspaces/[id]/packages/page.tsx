import { redirect, notFound } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { getWorkspaceByHash, getPackageRequest } from '@/app/actions/workspace.actions'
import { PackagesManager } from '../../_components/packages-manager'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function PackagesPage({ params }: PageProps) {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  const { id: hash } = await params

  const result = await getWorkspaceByHash(hash)

  if (!result.success || !result.data) {
    notFound()
  }

  const { workspace } = result.data
  const namespace = workspace.metadata?.namespace || 'default'
  const name = workspace.metadata?.name || ''

  // Fetch package request for initial data
  const pkgResult = await getPackageRequest(name, namespace)
  const packageRequest = pkgResult.success ? pkgResult.data : null

  return (
    <PackagesManager
      workspace={workspace as any}
      initialPackageRequest={packageRequest as any}
    />
  )
}
