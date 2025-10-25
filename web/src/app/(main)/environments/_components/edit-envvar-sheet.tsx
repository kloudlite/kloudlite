'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Loader2 } from 'lucide-react'
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
} from '@/components/ui/sheet'
import { setEnvVar } from '@/app/actions/environment-config'
import { toast } from 'sonner'
import type { EnvVar } from '@/types/environment'

interface EditEnvVarSheetProps {
  environmentId: string
  envVar: EnvVar
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function EditEnvVarSheet({
  environmentId,
  envVar,
  open,
  onOpenChange,
  onSuccess,
}: EditEnvVarSheetProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [value, setValue] = useState(envVar.value)

  // Update state when envVar changes
  useEffect(() => {
    setValue(envVar.value)
  }, [envVar])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!value.trim()) {
      toast.error('Please enter a value')
      return
    }

    startTransition(async () => {
      try {
        await setEnvVar(environmentId, envVar.key, value.trim(), envVar.type as 'config' | 'secret')

        toast.success('Environment variable updated successfully')
        onOpenChange(false)
        onSuccess?.()
        router.refresh()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to update environment variable')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full sm:max-w-xl">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Edit Envvar</SheetTitle>
            <SheetDescription>
              Update the value or type of this environment variable
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-6 overflow-y-auto p-4">
            <div className="space-y-2">
              <Label htmlFor="key">Key</Label>
              <Input
                id="key"
                value={envVar.key}
                disabled
                className="bg-muted font-mono text-sm"
              />
              <p className="text-muted-foreground text-sm">The key cannot be changed</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="value">Value</Label>
              <Input
                id="value"
                type={envVar.type === 'secret' ? 'password' : 'text'}
                placeholder={envVar.type === 'secret' ? '••••••••' : 'Enter new value'}
                value={value}
                onChange={(e) => setValue(e.target.value)}
                disabled={isPending}
                required
                className="font-mono text-sm"
              />
              <p className="text-muted-foreground text-sm">Update the envvar value</p>
            </div>

            <div className="space-y-2">
              <Label>Type</Label>
              <div className="rounded-lg border bg-muted p-3">
                <div className="flex items-center gap-2">
                  {envVar.type === 'config' ? (
                    <span className="inline-flex items-center rounded-full bg-info/10 px-2.5 py-0.5 text-xs font-medium text-info dark:bg-info/20">
                      Config
                    </span>
                  ) : (
                    <span className="inline-flex items-center rounded-full bg-purple-100 px-2.5 py-0.5 text-xs font-medium text-purple-800 dark:bg-purple-900/30 dark:text-purple-400">
                      Secret
                    </span>
                  )}
                  <span className="text-muted-foreground text-sm">
                    {envVar.type === 'config' ? 'Regular configuration variable' : 'Sensitive data'}
                  </span>
                </div>
              </div>
              <p className="text-muted-foreground text-sm">Type cannot be changed after creation</p>
            </div>
          </div>

          <SheetFooter className="flex-row justify-end gap-2 border-t p-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
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
