import { redirect } from 'next/navigation'
import type { PageProps } from '@/types/shared'

export default async function ConfigsPage({ params }: PageProps) {
  const { id } = await params
  redirect(`/environments/${id}/configs/envvars`)
}
