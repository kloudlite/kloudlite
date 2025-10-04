'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Edit2, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { setSecret } from '@/app/actions/environment-config'
import { toast } from 'sonner'

interface EditSecretSheetProps {
  environmentId: string
  secretKey: string
}

export function EditSecretSheet({ environmentId, secretKey }: EditSecretSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [value, setValue] = useState('')

  // Reset form when sheet opens
  useEffect(() => {
    if (open) {
      setValue('')
    }
  }, [open])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!value.trim()) {
      toast.error('Please enter a value')
      return
    }

    startTransition(async () => {
      try {
        const newData = { [secretKey]: value.trim() }
        await setSecret(environmentId, newData)

        toast.success('Secret updated successfully')
        setValue('')
        setOpen(false)
        router.refresh()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to update secret')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="sm">
          <Edit2 className="h-4 w-4" />
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-full sm:max-w-2xl">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Update Secret</SheetTitle>
            <SheetDescription>
              Update the encrypted secret value
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-4 overflow-y-auto p-4">
            <div className="space-y-2">
              <Label htmlFor="key">Key</Label>
              <Input
                id="key"
                value={secretKey}
                disabled
                className="font-mono text-sm bg-gray-50"
              />
              <p className="text-xs text-gray-500">Key cannot be changed</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="value">New Value</Label>
              <Input
                id="value"
                type="password"
                placeholder="Enter new secret value"
                value={value}
                onChange={(e) => setValue(e.target.value)}
                disabled={isPending}
                required
                className="font-mono text-sm"
              />
              <p className="text-xs text-gray-500">Current value cannot be displayed for security</p>
            </div>
          </div>

          <SheetFooter className="mt-auto">
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Update Secret
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
