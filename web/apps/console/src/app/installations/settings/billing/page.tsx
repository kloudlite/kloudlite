import { redirect } from 'next/navigation'

interface BillingPageProps {
  searchParams: Promise<{ installation?: string }>
}

export default async function BillingPage({ searchParams }: BillingPageProps) {
  const params = await searchParams
  const installationId = params.installation

  if (installationId) {
    redirect(`/installations/${installationId}/billing`)
  }

  redirect('/installations')
}
