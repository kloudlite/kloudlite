'use client'

import { useState } from 'react'
import { Badge } from '@kloudlite/ui'
import { ArrowDown, ArrowUp, ArrowUpDown } from 'lucide-react'
import { formatCurrency } from '@/lib/billing-utils'
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

type SortColumn = 'date' | 'amount' | 'status'
type SortDirection = 'asc' | 'desc'

const statusOrder: Record<string, number> = { issued: 0, paid: 1, expired: 2, cancelled: 3 }

function sortInvoices(invoices: Invoice[], column: SortColumn, direction: SortDirection): Invoice[] {
  return [...invoices].sort((a, b) => {
    let cmp = 0
    switch (column) {
      case 'date':
        cmp = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
        break
      case 'amount':
        cmp = a.amount - b.amount
        break
      case 'status':
        cmp = (statusOrder[a.status] ?? 99) - (statusOrder[b.status] ?? 99)
        break
    }
    return direction === 'asc' ? cmp : -cmp
  })
}

function SortIcon({ column, sortColumn, sortDirection }: { column: SortColumn; sortColumn: SortColumn; sortDirection: SortDirection }) {
  if (sortColumn !== column) return <ArrowUpDown className="h-3 w-3 opacity-40" />
  return sortDirection === 'asc'
    ? <ArrowUp className="h-3 w-3" />
    : <ArrowDown className="h-3 w-3" />
}

export function InvoiceHistory({ invoices }: InvoiceHistoryProps) {
  const [sortColumn, setSortColumn] = useState<SortColumn>('date')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')

  if (invoices.length === 0) return null

  const sorted = sortInvoices(invoices, sortColumn, sortDirection)

  const toggleSort = (column: SortColumn) => {
    if (sortColumn === column) {
      setSortDirection((d) => (d === 'asc' ? 'desc' : 'asc'))
    } else {
      setSortColumn(column)
      setSortDirection('desc')
    }
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-foreground/10">
            <th
              aria-sort={sortColumn === 'date' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}
              className="text-left py-2 pr-4 font-medium text-muted-foreground"
            >
              <button
                type="button"
                onClick={() => toggleSort('date')}
                className="inline-flex items-center gap-1 cursor-pointer select-none hover:text-foreground transition-colors"
              >
                Date <SortIcon column="date" sortColumn={sortColumn} sortDirection={sortDirection} />
              </button>
            </th>
            <th className="text-left py-2 pr-4 font-medium text-muted-foreground">Period</th>
            <th
              aria-sort={sortColumn === 'amount' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}
              className="text-right py-2 pr-4 font-medium text-muted-foreground"
            >
              <button
                type="button"
                onClick={() => toggleSort('amount')}
                className="ml-auto inline-flex items-center justify-end gap-1 cursor-pointer select-none hover:text-foreground transition-colors"
              >
                Amount <SortIcon column="amount" sortColumn={sortColumn} sortDirection={sortDirection} />
              </button>
            </th>
            <th
              aria-sort={sortColumn === 'status' ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}
              className="text-right py-2 font-medium text-muted-foreground"
            >
              <button
                type="button"
                onClick={() => toggleSort('status')}
                className="ml-auto inline-flex items-center justify-end gap-1 cursor-pointer select-none hover:text-foreground transition-colors"
              >
                Status <SortIcon column="status" sortColumn={sortColumn} sortDirection={sortDirection} />
              </button>
            </th>
          </tr>
        </thead>
        <tbody>
          {sorted.map((invoice) => (
            <tr key={invoice.id} className="border-b border-foreground/5">
              <td className="py-3 pr-4">
                {new Date(invoice.createdAt).toLocaleDateString()}
              </td>
              <td className="py-3 pr-4 text-muted-foreground">
                {invoice.billingStart && invoice.billingEnd
                  ? `${new Date(invoice.billingStart).toLocaleDateString()} \u2013 ${new Date(invoice.billingEnd).toLocaleDateString()}`
                  : '\u2014'}
              </td>
              <td className="py-3 pr-4 text-right font-medium">
                {formatCurrency(invoice.amount, invoice.currency)}
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
  )
}
