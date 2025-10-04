import { redirect } from 'next/navigation'

interface PageProps {
  params: {
    id: string
  }
}

export default function ResourcesPage({ params }: PageProps) {
  redirect(`/environments/${params.id}/resources/helmcharts`)
}
