import { redirect } from 'next/navigation'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function WorkspaceDetailPage({ params }: PageProps) {
  // Redirect to connect tab by default
  const { id } = await params
  redirect(`/workspaces/${id}/connect`)
}
