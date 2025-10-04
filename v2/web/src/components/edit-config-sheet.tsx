'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Edit2, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { setConfig, getConfig } from '@/app/actions/environment-config'
import { toast } from 'sonner'

interface EditConfigSheetProps {
  environmentId: string
  configKey: string
  configValue: string
}

export function EditConfigSheet({ environmentId, configKey, configValue }: EditConfigSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [key, setKey] = useState(configKey)
  const [value, setValue] = useState(configValue)

  // Reset form when sheet opens
  useEffect(() => {
    if (open) {
      setKey(configKey)
      setValue(configValue)
    }
  }, [open, configKey, configValue])

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
        // Get existing configs
        const existing = await getConfig(environmentId)
        const configsMap = { ...existing.data }

        // If key changed, remove old key
        if (key !== configKey) {
          delete configsMap[configKey]
        }

        configsMap[key.trim()] = value.trim()
        await setConfig(environmentId, configsMap)

        toast.success('Config updated successfully')
        setOpen(false)
        router.refresh()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to update config')
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
            <SheetTitle>Edit Config</SheetTitle>
            <SheetDescription>
              Update the configuration variable
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-4 overflow-y-auto p-4">
            <div className="space-y-2">
              <Label htmlFor="key">Key</Label>
              <Input
                id="key"
                value={key}
                onChange={(e) => setKey(e.target.value)}
                disabled={isPending}
                required
                className="font-mono text-sm"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="value">Value</Label>
              <Textarea
                id="value"
                value={value}
                onChange={(e) => setValue(e.target.value)}
                disabled={isPending}
                required
                rows={3}
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
              Save Changes
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
