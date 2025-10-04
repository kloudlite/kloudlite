'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import dynamic from 'next/dynamic'
import { FileCode, Loader2 } from 'lucide-react'
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
import { createComposition } from '@/app/actions/composition.actions'
import { toast } from 'sonner'

const CodeMirror = dynamic(() => import('@uiw/react-codemirror'), {
  ssr: false,
})

interface CreateCompositionSheetProps {
  namespace: string
  user: string
}

const defaultComposeContent = `services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
`

export function CreateCompositionSheet({ namespace, user }: CreateCompositionSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [name, setName] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [composeContent, setComposeContent] = useState(defaultComposeContent)
  const [yamlExtension, setYamlExtension] = useState<any>(null)

  useEffect(() => {
    import('@codemirror/lang-yaml').then((mod) => {
      setYamlExtension(mod.yaml())
    })
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      toast.error('Please enter a composition name')
      return
    }

    if (!displayName.trim()) {
      toast.error('Please enter a display name')
      return
    }

    if (!composeContent.trim()) {
      toast.error('Please enter compose content')
      return
    }

    startTransition(async () => {
      const result = await createComposition(
        namespace,
        {
          name: name.trim(),
          spec: {
            displayName: displayName.trim(),
            composeContent: composeContent,
            composeFormat: 'v3.8',
          },
        },
        user
      )

      if (result.success) {
        toast.success('Composition created successfully')
        setOpen(false)
        setName('')
        setDisplayName('')
        setComposeContent(defaultComposeContent)
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to create composition')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button size="sm" className="gap-2">
          <FileCode className="h-4 w-4" />
          Add Composition
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-full sm:max-w-2xl">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Create Composition</SheetTitle>
            <SheetDescription>
              Create a new composition from a docker-compose file
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-4 overflow-y-auto p-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                placeholder="my-composition"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isPending}
                required
              />
              <p className="text-xs text-gray-500">Kubernetes resource name (lowercase, alphanumeric, hyphens)</p>
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
              Create Composition
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
