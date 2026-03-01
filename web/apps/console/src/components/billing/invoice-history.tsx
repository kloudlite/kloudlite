import { Badge, Card, CardContent, CardHeader, CardTitle } from '@kloudlite/ui'
import { Receipt } from 'lucide-react'
import type { Invoice } from '@/lib/console/storage'

interface InvoiceHistoryProps {
  invoices: Invoice[]
}

const invoiceStatusColors: Record<string, string> = {
  paid: 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20',
  issued: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20',
  expired: 'bg-gray-500/10 text-gray-700 dark:text-gray-400 border-gray-500/20',
  cancelled: 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
}

export function InvoiceHistory({ invoices }: InvoiceHistoryProps) {
  if (invoices.length === 0) return null

  return (
    <Card className="border-foreground/10">
      <CardHeader className="pb-4">
        <CardTitle className="text-lg flex items-center gap-2">
          <Receipt className="h-5 w-5" />
          Invoice History
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-foreground/10">
                <th className="text-left py-2 pr-4 font-medium text-muted-foreground">Date</th>
                <th className="text-left py-2 pr-4 font-medium text-muted-foreground">Period</th>
                <th className="text-right py-2 pr-4 font-medium text-muted-foreground">Amount</th>
                <th className="text-right py-2 font-medium text-muted-foreground">Status</th>
              </tr>
            </thead>
            <tbody>
              {invoices.map((invoice) => (
                <tr key={invoice.id} className="border-b border-foreground/5">
                  <td className="py-3 pr-4">
                    {new Date(invoice.createdAt).toLocaleDateString()}
                  </td>
                  <td className="py-3 pr-4 text-muted-foreground">
                    {invoice.billingStart && invoice.billingEnd
                      ? `${new Date(invoice.billingStart).toLocaleDateString()} \u2014 ${new Date(invoice.billingEnd).toLocaleDateString()}`
                      : '\u2014'}
                  </td>
                  <td className="py-3 pr-4 text-right font-medium">
                    ${(invoice.amount / 100).toFixed(2)} {invoice.currency}
                  </td>
                  <td className="py-3 text-right">
                    <Badge variant="outline" className={invoiceStatusColors[invoice.status]}>
                      {invoice.status}
                    </Badge>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  )
}
