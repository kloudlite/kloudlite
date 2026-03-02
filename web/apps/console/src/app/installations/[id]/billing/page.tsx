import { redirect } from 'next/navigation'

interface BillingPageProps {
  params: Promise<{ id: string }>
}

export default async function BillingPage({ params }: BillingPageProps) {
  const { id } = await params
  redirect(`/installations/${id}/billing/subscription`)
}
