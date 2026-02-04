import { redirect } from 'next/navigation'
import type { PageProps } from '@/types/shared'

export default async function EnvironmentDetailPage({ params }: PageProps) {
  // Redirect to services tab by default
  const { id } = await params
  redirect(`/environments/${id}/services`)
}
