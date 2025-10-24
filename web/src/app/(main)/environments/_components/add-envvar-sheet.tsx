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
import { createEnvVar } from '@/app/actions/environment-config'
import { toast } from 'sonner'

interface AddEnvVarSheetProps {
  environmentId: string
  onSuccess?: () => void
}

export function AddEnvVarSheet({ environmentId, onSuccess }: AddEnvVarSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [key, setKey] = useState('')
  const [value, setValue] = useState('')
  const [type, setType] = useState<'config' | 'secret'>('config')

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
        await createEnvVar(environmentId, key.trim(), value.trim(), type)

        toast.success('Environment variable added successfully')
        setKey('')
        setValue('')
        setType('config')
        setOpen(false)
        onSuccess?.()
        router.refresh()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to add environment variable')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button size="sm">
          <Plus className="mr-2 h-4 w-4" />
          Add Variable
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-full sm:max-w-xl">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Add Envvar</SheetTitle>
            <SheetDescription>Add a new configuration or secret envvar</SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-6 overflow-y-auto p-4">
            <div className="space-y-2">
              <Label htmlFor="key">Key</Label>
              <Input
                id="key"
                placeholder="DATABASE_URL"
                value={key}
                onChange={(e) => setKey(e.target.value)}
                disabled={isPending}
                required
                className="font-mono text-sm"
              />
              <p className="text-muted-foreground text-sm">The envvar name</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="value">Value</Label>
              <Input
                id="value"
                type={type === 'secret' ? 'password' : 'text'}
                placeholder={type === 'secret' ? '••••••••' : 'postgresql://...'}
                value={value}
                onChange={(e) => setValue(e.target.value)}
                disabled={isPending}
                required
                className="font-mono text-sm"
              />
              <p className="text-muted-foreground text-sm">The envvar value</p>
            </div>

            <div className="space-y-3">
              <Label>Type</Label>
              <div className="space-y-2">
                <label className="hover:bg-accent/50 flex cursor-pointer items-center space-x-3 rounded-lg border p-3">
                  <input
                    type="radio"
                    name="type"
                    value="config"
                    checked={type === 'config'}
                    onChange={(e) => setType(e.target.value as 'config' | 'secret')}
                    className="h-4 w-4 text-info"
                  />
                  <div className="flex-1">
                    <div className="font-medium">Config</div>
                    <div className="text-muted-foreground text-sm">
                      Regular configuration variable (visible in list)
                    </div>
                  </div>
                </label>
                <label className="hover:bg-accent/50 flex cursor-pointer items-center space-x-3 rounded-lg border p-3">
                  <input
                    type="radio"
                    name="type"
                    value="secret"
                    checked={type === 'secret'}
                    onChange={(e) => setType(e.target.value as 'config' | 'secret')}
                    className="h-4 w-4 text-info"
                  />
                  <div className="flex-1">
                    <div className="font-medium">Secret</div>
                    <div className="text-muted-foreground text-sm">
                      Sensitive data (value hidden in list)
                    </div>
                  </div>
                </label>
              </div>
            </div>
          </div>

          <SheetFooter className="flex-row justify-end gap-2 border-t p-4">
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
              Add Variable
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
