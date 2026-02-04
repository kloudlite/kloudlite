import { redirect } from 'next/navigation'
import type { PageProps } from '@/types/shared'

export default async function SettingsPage({ params }: PageProps) {
  const { id } = await params
  redirect(`/environments/${id}/settings/access`)
}
