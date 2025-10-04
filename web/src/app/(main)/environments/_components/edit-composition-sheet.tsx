'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import dynamic from 'next/dynamic'
import { Pencil, Loader2 } from 'lucide-react'
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
import { updateComposition } from '@/app/actions/composition.actions'
import { toast } from 'sonner'
import type { Composition } from '@/types/composition'

const CodeMirror = dynamic(() => import('@uiw/react-codemirror'), {
  ssr: false,
})

interface EditCompositionSheetProps {
  composition: Composition
  namespace: string
  user: string
}

export function EditCompositionSheet({ composition, namespace, user }: EditCompositionSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [displayName, setDisplayName] = useState(composition.spec.displayName)
  const [composeContent, setComposeContent] = useState(composition.spec.composeContent)
  const [yamlExtension, setYamlExtension] = useState<any>(null)

  useEffect(() => {
    import('@codemirror/lang-yaml').then((mod) => {
      setYamlExtension(mod.yaml())
    })
  }, [])

  // Reset form when composition changes or sheet opens
  useEffect(() => {
    if (open) {
      setDisplayName(composition.spec.displayName)
      setComposeContent(composition.spec.composeContent)
    }
  }, [open, composition])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!displayName.trim()) {
      toast.error('Please enter a display name')
      return
    }

    if (!composeContent.trim()) {
      toast.error('Please enter compose content')
      return
    }

    startTransition(async () => {
      const result = await updateComposition(
        namespace,
        composition.metadata.name,
        {
          spec: {
            displayName: displayName.trim(),
            composeContent: composeContent,
            composeFormat: composition.spec.composeFormat || 'v3.8',
          },
        },
        user
      )

      if (result.success) {
        toast.success('Composition updated successfully')
        setOpen(false)
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to update composition')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
          <Pencil className="h-4 w-4" />
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-full sm:max-w-2xl">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Edit Composition</SheetTitle>
            <SheetDescription>
              Update the composition configuration
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-4 overflow-y-auto p-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={composition.metadata.name}
                disabled
                className="bg-gray-50"
              />
              <p className="text-xs text-gray-500">Name cannot be changed</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="display-name">Display Name</Label>
              <Input
                id="display-name"
                placeholder="My Composition"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                disabled={isPending}
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="compose-content">Docker Compose Content</Label>
              <div className="rounded-md border">
                {yamlExtension ? (
                  <CodeMirror
                    value={composeContent}
                    height="400px"
                    extensions={[yamlExtension]}
                    onChange={(value) => setComposeContent(value)}
                    className="text-sm"
                  />
                ) : (
                  <div className="h-[400px] flex items-center justify-center text-gray-400">
                    Loading editor...
                  </div>
                )}
              </div>
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
              Update Composition
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
