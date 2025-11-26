'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Plus, Loader2 } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import { Textarea } from '@kloudlite/ui'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@kloudlite/ui'
import { setFile } from '@/app/actions/environment-config'
import { toast } from 'sonner'

interface AddFileSheetProps {
  environmentId: string
}

export function AddFileSheet({ environmentId }: AddFileSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [filename, setFilename] = useState('')
  const [content, setContent] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!filename.trim()) {
      toast.error('Please enter a filename')
      return
    }

    if (!content.trim()) {
      toast.error('Please enter file content')
      return
    }

    startTransition(async () => {
      try {
        await setFile(environmentId, filename.trim(), content.trim())

        toast.success('File added successfully')
        setFilename('')
        setContent('')
        setOpen(false)
        router.refresh()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to add file')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button size="sm">
          <Plus className="mr-2 h-4 w-4" />
          Add File
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-full sm:max-w-2xl">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Add File</SheetTitle>
            <SheetDescription>Add a new configuration file</SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-4 overflow-y-auto p-4">
            <div className="space-y-2">
              <Label htmlFor="filename">Filename</Label>
              <Input
                id="filename"
                placeholder="nginx.conf"
                value={filename}
                onChange={(e) => setFilename(e.target.value)}
                disabled={isPending}
                required
                className="font-mono text-sm"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="content">Content</Label>
              <Textarea
                id="content"
                placeholder="Enter file content..."
                value={content}
                onChange={(e) => setContent(e.target.value)}
                disabled={isPending}
                required
                rows={12}
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
              Add File
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
