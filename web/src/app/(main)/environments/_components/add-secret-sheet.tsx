'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Plus, Loader2 } from 'lucide-react'
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

interface AddSecretSheetProps {
  environmentId: string
}

export function AddSecretSheet({ environmentId }: AddSecretSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [key, setKey] = useState('')
  const [value, setValue] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!key.trim()) {
      toast.error('Please enter a key')
      return
    }

    if (!value.trim()) {
      toast.error('Please enter a value')
      return
    }

    startTransition(async () => {
      try {
        const newData = { [key.trim()]: value.trim() }
        await setSecret(environmentId, newData)

        toast.success('Secret added successfully')
        setKey('')
        setValue('')
        setOpen(false)
        router.refresh()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to add secret')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button size="sm">
          <Plus className="h-4 w-4 mr-2" />
          Add Secret
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-full sm:max-w-2xl">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Add Secret</SheetTitle>
            <SheetDescription>
              Add a new encrypted secret
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-4 overflow-y-auto p-4">
            <div className="space-y-2">
              <Label htmlFor="key">Key</Label>
              <Input
                id="key"
                placeholder="DB_PASSWORD"
                value={key}
                onChange={(e) => setKey(e.target.value)}
                disabled={isPending}
                required
                className="font-mono text-sm"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="value">Value</Label>
              <Input
                id="value"
                type="password"
                placeholder="Enter secret value"
                value={value}
                onChange={(e) => setValue(e.target.value)}
                disabled={isPending}
                required
                className="font-mono text-sm"
              />
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
              Add Secret
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
