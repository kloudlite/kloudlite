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
import { setFile, getFile, deleteFile } from '@/app/actions/environment-config'
import { toast } from 'sonner'

interface EditFileSheetProps {
  environmentId: string
  filename: string
}

export function EditFileSheet({ environmentId, filename }: EditFileSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [loadingFile, setLoadingFile] = useState(false)
  const [editFilename, setEditFilename] = useState(filename)
  const [content, setContent] = useState('')

  // Load file content when sheet opens
  useEffect(() => {
    if (open) {
      setEditFilename(filename)
      setLoadingFile(true)
      setContent('')

      getFile(environmentId, filename)
        .then((response) => {
          setContent(response.content)
        })
        .catch((err) => {
          toast.error(err instanceof Error ? err.message : 'Failed to load file content')
          setOpen(false)
        })
        .finally(() => {
          setLoadingFile(false)
        })
    }
  }, [open, environmentId, filename])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!editFilename.trim()) {
      toast.error('Please enter a filename')
      return
    }

    if (!content.trim()) {
      toast.error('Please enter file content')
      return
    }

    startTransition(async () => {
      try {
        // If filename changed, delete the old one and create new one
        if (editFilename !== filename) {
          await deleteFile(environmentId, filename)
        }

        await setFile(environmentId, editFilename.trim(), content.trim())

        toast.success('File updated successfully')
        setOpen(false)
        router.refresh()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to update file')
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
            <SheetTitle>Edit File</SheetTitle>
            <SheetDescription>Update the configuration file</SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-4 overflow-y-auto p-4">
            <div className="space-y-2">
              <Label htmlFor="filename">Filename</Label>
              <Input
                id="filename"
                value={editFilename}
                onChange={(e) => setEditFilename(e.target.value)}
                disabled={isPending || loadingFile}
                required
                className="font-mono text-sm"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="content">Content</Label>
              {loadingFile ? (
                <div className="flex items-center justify-center rounded-md border border-gray-300 py-12">
                  <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
                </div>
              ) : (
                <Textarea
                  id="content"
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  disabled={isPending}
                  required
                  rows={12}
                  className="font-mono text-sm"
                />
              )}
            </div>
          </div>

          <SheetFooter className="mt-auto">
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
              disabled={isPending || loadingFile}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isPending || loadingFile}>
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Save Changes
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
