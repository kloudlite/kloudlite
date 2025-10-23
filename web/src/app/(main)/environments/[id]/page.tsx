import { redirect } from 'next/navigation'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function EnvironmentDetailPage({ params }: PageProps) {
  // Redirect to services tab by default
  const { id } = await params
  redirect(`/environments/${id}/services`)
}
