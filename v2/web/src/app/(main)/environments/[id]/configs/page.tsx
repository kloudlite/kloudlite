import { redirect } from 'next/navigation'

interface PageProps {
  params: {
    id: string
  }
}

export default function ConfigsPage({ params }: PageProps) {
  redirect(`/environments/${params.id}/configs/configmaps`)
}
