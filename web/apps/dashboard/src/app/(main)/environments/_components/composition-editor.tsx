'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import dynamic from 'next/dynamic'
import { Loader2, Save, Pencil } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@kloudlite/ui'
import { updateEnvironmentCompose } from '@/app/actions/environment.actions'
import { toast } from 'sonner'
import type { Extension } from '@codemirror/state'

const CodeMirror = dynamic(() => import('@uiw/react-codemirror'), {
  ssr: false,
  loading: () => (
    <div className="flex h-[500px] items-center justify-center bg-muted/30 rounded-md">
      <div className="flex flex-col items-center gap-2">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        <span className="text-sm text-muted-foreground">Loading editor...</span>
      </div>
    </div>
  ),
})

interface CompositionEditorProps {
  environmentName: string
  composeContent: string | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

const defaultComposeContent = `services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
`

export function CompositionEditor({
  environmentName,
  composeContent: initialComposeContent,
  open,
  onOpenChange,
}: CompositionEditorProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [composeContent, setComposeContent] = useState(
    initialComposeContent || defaultComposeContent,
  )
  const [yamlExtension, setYamlExtension] = useState<Extension | null>(null)

  useEffect(() => {
    import('@codemirror/lang-yaml')
      .then((mod) => {
        setYamlExtension(mod.yaml())
      })
      .catch((err) => {
        console.error('Failed to load YAML extension:', err)
        setYamlExtension(null)
      })
  }, [])

  // Update compose content when prop changes or sheet opens
  useEffect(() => {
    if (open && initialComposeContent) {
      setComposeContent(initialComposeContent)
    }
  }, [open, initialComposeContent])

  const handleSave = async () => {
    startTransition(async () => {
      // Update compose content in environment
      const result = await updateEnvironmentCompose(environmentName, {
        displayName: 'Main Composition',
        composeContent: composeContent,
        composeFormat: 'v3.8',
      })

      if (result.success) {
        toast.success('Composition saved successfully')
        onOpenChange(false)
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to save composition')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetTrigger asChild>
        <Button size="sm" variant="outline" className="gap-2">
          <Pencil className="h-4 w-4" />
          Composition
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-full sm:max-w-2xl">
        <div className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Edit Composition</SheetTitle>
            <SheetDescription>Define your services using Docker Compose format</SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-4 overflow-y-auto p-4">
            <div className="space-y-2">
              <Label htmlFor="compose-content">Docker Compose YAML</Label>
              <div className="rounded-md border">
                {yamlExtension ? (
                  <CodeMirror
                    value={composeContent}
                    height="500px"
                    extensions={[yamlExtension]}
                    onChange={(value) => setComposeContent(value)}
                    className="text-sm"
                  />
                ) : (
                  <div className="text-muted-foreground flex h-[500px] items-center justify-center">
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
              onClick={() => onOpenChange(false)}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={isPending}>
              {isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  Save Changes
                </>
              )}
            </Button>
          </SheetFooter>
        </div>
      </SheetContent>
    </Sheet>
  )
}
