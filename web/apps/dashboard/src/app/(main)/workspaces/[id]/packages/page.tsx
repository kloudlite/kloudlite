import { redirect, notFound } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { getPackageRequest } from '@/app/actions/workspace.actions'
import { getWorkspaceData } from '../workspace-data'
import { PackagesList } from '../../_components/packages-list'
import type { PageProps } from '@/types/shared'

export default async function PackagesPage({ params }: PageProps) {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  const { id: hash } = await params

  const result = await getWorkspaceData(hash)

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
    <PackagesList
      workspace={workspace as any}
      initialPackageRequest={packageRequest as any}
    />
  )
}
