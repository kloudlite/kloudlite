import { redirect } from 'next/navigation'

interface PageProps {
  params: {
    id: string
  }
}

export default async function ConfigsPage({ params }: PageProps) {
  const { id } = await params
  redirect(`/environments/${id}/configs/envvars`)
}
