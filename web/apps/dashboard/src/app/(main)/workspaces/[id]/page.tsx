import { redirect } from 'next/navigation'
import type { PageProps } from '@/types/shared'

export default async function WorkspaceDetailPage({ params }: PageProps) {
  // Redirect to overview tab by default
  const { id } = await params
  redirect(`/workspaces/${id}/overview`)
}
