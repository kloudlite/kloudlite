import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { InvoiceHistory } from '@/components/billing/invoice-history'
import { getInvoicesByInstallation } from '@/lib/console/storage'

interface InvoicesPageProps {
  params: Promise<{ id: string }>
}

export default async function InvoicesPage({ params }: InvoicesPageProps) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  const invoices = await getInvoicesByInstallation(id)

  if (invoices.length === 0) {
    return (
      <p className="text-muted-foreground text-sm py-8 text-center">No invoices yet.</p>
    )
  }

  return <InvoiceHistory invoices={invoices} />
}
