'use client'

import { useState } from 'react'
import { ExternalLink, Plus, Loader2, Wallet } from 'lucide-react'
import {
  Button,
  Input,
  Label,
  Switch,
  Badge,
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

function cleanDescription(desc: string | null): string {
  if (!desc) return '-'
  // Remove Stripe session/invoice IDs from descriptions
  return desc
    .replace(/\s*cs_test_\S+/g, '')
    .replace(/\s*in_\S+/g, '')
    .trim() || desc
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
      <div className="border border-foreground/10 rounded-lg p-8 text-center">
        <p className="text-muted-foreground text-sm">
          No billing account found. Create an installation to get started.
        </p>
      </div>
    )
  }

  const handleTopupSubmit = async () => {
    const amount = parseFloat(topupAmount)
    if (isNaN(amount) || amount <= 0) return
    await handleTopup(amount, '/installations/settings/billing')
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
      {/* Balance + Actions */}
      <div className="border border-foreground/10 rounded-lg bg-background">
        <div className="border-b border-foreground/10 px-6 py-4 flex items-center justify-between">
          <h3 className="font-medium text-foreground">Credit Balance</h3>
          <div className="flex items-center gap-2">
            {isOwner && (
              <>
                <Button variant="outline" size="sm" onClick={handleManageBilling} disabled={loading}>
                  Manage Billing
                  <ExternalLink className="ml-1.5 h-3 w-3" />
                </Button>
                <Dialog open={topupDialogOpen} onOpenChange={setTopupDialogOpen}>
                  <DialogTrigger asChild>
                    <Button variant="outline" size="sm">
                      <Plus className="mr-1 h-3.5 w-3.5" />
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
              </>
            )}
          </div>
        </div>
        <div className="px-6 py-5">
          <div className="flex items-baseline gap-3">
            <Wallet className="h-5 w-5 text-muted-foreground" />
            <span className={`text-3xl font-bold tabular-nums ${getBalanceColor(account)}`}>
              {formatCurrency(account.balance)}
            </span>
          </div>
          {account.negativeBalanceFlagged && (
            <p className="mt-2 text-sm text-red-600 dark:text-red-400">
              Your balance is negative. Please add credits to avoid service interruption.
            </p>
          )}
          {account.lowBalanceWarning && !account.negativeBalanceFlagged && (
            <p className="mt-2 text-sm text-yellow-600 dark:text-yellow-400">
              Your balance is running low. Consider adding credits.
            </p>
          )}
        </div>
      </div>

      {/* Active Usage */}
      {activePeriods.length > 0 && (
        <div className="border border-foreground/10 rounded-lg bg-background">
          <div className="border-b border-foreground/10 px-6 py-4">
            <h3 className="font-medium text-foreground">Active Usage</h3>
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Resource</TableHead>
                <TableHead>Started</TableHead>
                <TableHead className="text-right">Rate</TableHead>
                <TableHead className="text-right">Running Cost</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {activePeriods.map((period) => (
                <TableRow key={period.id}>
                  <TableCell>
                    <span className="font-medium text-sm">{period.resourceType}</span>
                    <span className="text-muted-foreground text-xs ml-2 font-mono">{period.resourceId.slice(0, 8)}...</span>
                  </TableCell>
                  <TableCell className="text-sm">{formatRelativeTime(period.startedAt)}</TableCell>
                  <TableCell className="text-right text-sm tabular-nums">
                    {formatCurrency(period.hourlyRate)}/hr
                  </TableCell>
                  <TableCell className="text-right text-sm tabular-nums">
                    {formatCurrency(getRunningCost(period.startedAt, period.hourlyRate))}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Auto Top-Up (owner only) */}
      {isOwner && (
        <div className="border border-foreground/10 rounded-lg bg-background">
          <div className={`px-6 py-4 flex items-center justify-between ${autoTopupEnabled ? 'border-b border-foreground/10' : ''}`}>
            <div>
              <h3 className="font-medium text-foreground">Auto Top-Up</h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                Automatically add credits when your balance drops below a threshold
              </p>
            </div>
            <Switch
              checked={autoTopupEnabled}
              onCheckedChange={setAutoTopupEnabled}
            />
          </div>
          {autoTopupEnabled && (
            <div className="px-6 py-4 space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="grid gap-1.5">
                  <Label htmlFor="threshold" className="text-sm">When balance drops below</Label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground text-sm">$</span>
                    <Input
                      id="threshold"
                      type="number"
                      min="1"
                      step="1"
                      value={autoTopupThreshold}
                      onChange={(e) => setAutoTopupThreshold(e.target.value)}
                      className="pl-7"
                    />
                  </div>
                </div>
                <div className="grid gap-1.5">
                  <Label htmlFor="topup-amount-auto" className="text-sm">Add this amount</Label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground text-sm">$</span>
                    <Input
                      id="topup-amount-auto"
                      type="number"
                      min="5"
                      step="5"
                      value={autoTopupAmount}
                      onChange={(e) => setAutoTopupAmount(e.target.value)}
                      className="pl-7"
                    />
                  </div>
                </div>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={handleSaveAutoTopup}
                disabled={savingAutoTopup}
              >
                {savingAutoTopup && <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />}
                Save Settings
              </Button>
            </div>
          )}
        </div>
      )}

      {/* Transaction History */}
      {transactions.length > 0 && (
        <div className="border border-foreground/10 rounded-lg bg-background">
          <div className="border-b border-foreground/10 px-6 py-4">
            <h3 className="font-medium text-foreground">Transaction History</h3>
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Date</TableHead>
                <TableHead>Type</TableHead>
                <TableHead className="text-right">Amount</TableHead>
                <TableHead>Description</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {transactions.slice(0, 50).map((tx) => (
                <TableRow key={tx.id}>
                  <TableCell className="whitespace-nowrap text-sm">
                    {formatDate(tx.createdAt)}
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant={
                        tx.type === 'topup' ? 'secondary' :
                        tx.type === 'usage_debit' ? 'destructive' : 'outline'
                      }
                      className="text-xs"
                    >
                      {tx.type === 'topup' ? 'Credit' : tx.type === 'usage_debit' ? 'Debit' : 'Adjustment'}
                    </Badge>
                  </TableCell>
                  <TableCell
                    className={`text-right tabular-nums text-sm ${
                      tx.type === 'topup'
                        ? 'text-green-600 dark:text-green-400'
                        : tx.type === 'usage_debit'
                          ? 'text-red-600 dark:text-red-400'
                          : ''
                    }`}
                  >
                    {tx.type === 'topup' ? '+' : tx.type === 'usage_debit' ? '-' : ''}
                    {formatCurrency(Math.abs(tx.amount))}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">
                    {cleanDescription(tx.description)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {!isOwner && (
        <p className="text-muted-foreground text-center text-sm py-4">
          Only the organization owner can manage billing settings.
        </p>
      )}
    </div>
  )
}
