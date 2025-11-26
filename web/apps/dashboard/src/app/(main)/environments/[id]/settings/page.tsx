import { redirect } from 'next/navigation'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function SettingsPage({ params }: PageProps) {
  const { id } = await params
  redirect(`/environments/${id}/settings/general`)
}
