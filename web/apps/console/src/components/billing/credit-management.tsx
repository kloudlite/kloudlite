'use client'

import { useState } from 'react'
import { ExternalLink, Plus, Loader2 } from 'lucide-react'
import {
  Button,
  Input,
  Label,
  Switch,
  Badge,
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from '@kloudlite/ui'
import { useCredits } from '@/hooks/use-credits'

interface CreditManagementProps {
  orgId: string
  isOwner: boolean
}

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
  }).format(amount)
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatRelativeTime(iso: string): string {
  const ms = Date.now() - Date.parse(iso)
  const minutes = Math.floor(ms / 60000)
  if (minutes < 1) return 'just now'
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

function getRunningCost(startedAt: string, hourlyRate: number): number {
  const ms = Date.now() - Date.parse(startedAt)
  return (ms / 3600000) * hourlyRate
}

function getBalanceColor(account: {
  lowBalanceWarning: boolean
  negativeBalanceFlagged: boolean
  balance: number
}): string {
  if (account.negativeBalanceFlagged || account.balance < 0) {
    return 'text-red-600 dark:text-red-400'
  }
  if (account.lowBalanceWarning) {
    return 'text-yellow-600 dark:text-yellow-400'
  }
  return 'text-green-600 dark:text-green-400'
}

function getTransactionBadgeVariant(
  type: 'topup' | 'usage_debit' | 'adjustment',
): 'default' | 'destructive' | 'secondary' | 'outline' {
  switch (type) {
    case 'topup':
      return 'default'
    case 'usage_debit':
      return 'destructive'
    case 'adjustment':
      return 'secondary'
  }
}

function getTransactionLabel(type: 'topup' | 'usage_debit' | 'adjustment'): string {
  switch (type) {
    case 'topup':
      return 'Top-up'
    case 'usage_debit':
      return 'Debit'
    case 'adjustment':
      return 'Adjustment'
  }
}

export function CreditManagement({ orgId, isOwner }: CreditManagementProps) {
  const { loading, data, handleTopup, handleManageBilling, handleUpdateAutoTopup } =
    useCredits(orgId)

  const [topupAmount, setTopupAmount] = useState('25')
  const [topupDialogOpen, setTopupDialogOpen] = useState(false)

  const [autoTopupEnabled, setAutoTopupEnabled] = useState(
    data?.account?.autoTopupEnabled ?? false,
  )
  const [autoTopupThreshold, setAutoTopupThreshold] = useState(
    String(data?.account?.autoTopupThreshold ?? '10'),
  )
  const [autoTopupAmount, setAutoTopupAmount] = useState(
    String(data?.account?.autoTopupAmount ?? '25'),
  )
  const [savingAutoTopup, setSavingAutoTopup] = useState(false)

  // Sync local state when data loads
  const account = data?.account
  const transactions = data?.transactions ?? []
  const activePeriods = data?.activePeriods ?? []

  if (loading && !data) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!account) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground text-sm">
            No billing account found. Create an installation to get started.
          </p>
        </CardContent>
      </Card>
    )
  }

  const handleTopupSubmit = async () => {
    const amount = parseFloat(topupAmount)
    if (isNaN(amount) || amount <= 0) return
    await handleTopup(amount)
    setTopupDialogOpen(false)
  }

  const handleSaveAutoTopup = async () => {
    setSavingAutoTopup(true)
    await handleUpdateAutoTopup(
      autoTopupEnabled,
      autoTopupEnabled ? parseFloat(autoTopupThreshold) : undefined,
      autoTopupEnabled ? parseFloat(autoTopupAmount) : undefined,
    )
    setSavingAutoTopup(false)
  }

  return (
    <div className="space-y-6">
      {/* Credit Balance Card */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Credit Balance</CardTitle>
          {isOwner && (
            <Dialog open={topupDialogOpen} onOpenChange={setTopupDialogOpen}>
              <DialogTrigger asChild>
                <Button size="sm">
                  <Plus className="mr-1.5 h-3.5 w-3.5" />
                  Add Credits
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Add Credits</DialogTitle>
                  <DialogDescription>
                    Enter the amount you&apos;d like to add to your credit balance.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="topup-amount">Amount (USD)</Label>
                    <Input
                      id="topup-amount"
                      type="number"
                      min="5"
                      step="5"
                      value={topupAmount}
                      onChange={(e) => setTopupAmount(e.target.value)}
                      placeholder="25.00"
                    />
                  </div>
                </div>
                <DialogFooter>
                  <DialogClose asChild>
                    <Button variant="outline">Cancel</Button>
                  </DialogClose>
                  <Button
                    onClick={handleTopupSubmit}
                    disabled={loading || !topupAmount || parseFloat(topupAmount) <= 0}
                  >
                    {loading && <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />}
                    Add {topupAmount ? formatCurrency(parseFloat(topupAmount)) : '$0.00'}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          )}
        </CardHeader>
        <CardContent>
          <div className={`text-3xl font-bold ${getBalanceColor(account)}`}>
            {formatCurrency(account.balance)}
          </div>
          {account.negativeBalanceFlagged && (
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">
              Your balance is negative. Please add credits to avoid service interruption.
            </p>
          )}
          {account.lowBalanceWarning && !account.negativeBalanceFlagged && (
            <p className="mt-1 text-sm text-yellow-600 dark:text-yellow-400">
              Your balance is running low. Consider adding credits.
            </p>
          )}
        </CardContent>
      </Card>

      {/* Active Usage */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium">Active Usage</CardTitle>
        </CardHeader>
        <CardContent>
          {activePeriods.length === 0 ? (
            <p className="text-sm text-muted-foreground">No active resources</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Resource Type</TableHead>
                  <TableHead>Resource ID</TableHead>
                  <TableHead>Started</TableHead>
                  <TableHead className="text-right">Hourly Rate</TableHead>
                  <TableHead className="text-right">Running Cost</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {activePeriods.map((period) => (
                  <TableRow key={period.id}>
                    <TableCell className="font-medium">{period.resourceType}</TableCell>
                    <TableCell className="font-mono text-xs">{period.resourceId}</TableCell>
                    <TableCell>{formatRelativeTime(period.startedAt)}</TableCell>
                    <TableCell className="text-right">
                      {formatCurrency(period.hourlyRate)}/hr
                    </TableCell>
                    <TableCell className="text-right">
                      {formatCurrency(getRunningCost(period.startedAt, period.hourlyRate))}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Auto Top-Up Settings (owner only) */}
      {isOwner && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Auto Top-Up</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <Label htmlFor="auto-topup-toggle">Enable auto top-up</Label>
                <p className="text-xs text-muted-foreground">
                  Automatically add credits when your balance drops below a threshold.
                </p>
              </div>
              <Switch
                id="auto-topup-toggle"
                checked={autoTopupEnabled}
                onCheckedChange={setAutoTopupEnabled}
              />
            </div>
            {autoTopupEnabled && (
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="grid gap-2">
                  <Label htmlFor="threshold">Top up when balance drops below ($)</Label>
                  <Input
                    id="threshold"
                    type="number"
                    min="1"
                    step="1"
                    value={autoTopupThreshold}
                    onChange={(e) => setAutoTopupThreshold(e.target.value)}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="topup-amount-auto">Top up amount ($)</Label>
                  <Input
                    id="topup-amount-auto"
                    type="number"
                    min="5"
                    step="5"
                    value={autoTopupAmount}
                    onChange={(e) => setAutoTopupAmount(e.target.value)}
                  />
                </div>
              </div>
            )}
            <Button
              size="sm"
              onClick={handleSaveAutoTopup}
              disabled={savingAutoTopup}
            >
              {savingAutoTopup && <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />}
              Save Settings
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Transaction History */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium">Transaction History</CardTitle>
        </CardHeader>
        <CardContent>
          {transactions.length === 0 ? (
            <p className="text-sm text-muted-foreground">No transactions yet</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Date</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Description</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {transactions.slice(0, 50).map((tx) => (
                  <TableRow key={tx.id}>
                    <TableCell className="whitespace-nowrap">
                      {formatDate(tx.createdAt)}
                    </TableCell>
                    <TableCell>
                      <Badge variant={getTransactionBadgeVariant(tx.type)}>
                        {getTransactionLabel(tx.type)}
                      </Badge>
                    </TableCell>
                    <TableCell
                      className={
                        tx.type === 'topup'
                          ? 'text-green-600 dark:text-green-400'
                          : tx.type === 'usage_debit'
                            ? 'text-red-600 dark:text-red-400'
                            : ''
                      }
                    >
                      {tx.type === 'topup' ? '+' : tx.type === 'usage_debit' ? '-' : ''}
                      {formatCurrency(Math.abs(tx.amount))}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {tx.description ?? '-'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Manage Billing */}
      {isOwner && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Billing Management</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-wrap gap-3">
            <Button variant="outline" size="sm" onClick={handleManageBilling} disabled={loading}>
              Manage Payment Methods
              <ExternalLink className="ml-1.5 h-3.5 w-3.5" />
            </Button>
            <Button variant="outline" size="sm" onClick={handleManageBilling} disabled={loading}>
              View Invoices
              <ExternalLink className="ml-1.5 h-3.5 w-3.5" />
            </Button>
          </CardContent>
        </Card>
      )}

      {!isOwner && (
        <p className="text-muted-foreground text-center text-sm py-4">
          Only the organization owner can manage billing settings.
        </p>
      )}
    </div>
  )
}
