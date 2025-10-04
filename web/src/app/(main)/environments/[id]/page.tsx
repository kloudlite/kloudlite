import { redirect } from 'next/navigation'

interface PageProps {
  params: {
    id: string
  }
}

export default function EnvironmentDetailPage({ params }: PageProps) {
  // Redirect to resources tab by default
  redirect(`/environments/${params.id}/resources`)
}