import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { BillingContent } from '@/components/billing-content'

export default async function BillingPage() {
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  return <BillingContent />
}
