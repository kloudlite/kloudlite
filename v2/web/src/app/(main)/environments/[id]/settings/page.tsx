import { redirect } from 'next/navigation'

interface PageProps {
  params: {
    id: string
  }
}

export default function SettingsPage({ params }: PageProps) {
  redirect(`/environments/${params.id}/settings/general`)
}
