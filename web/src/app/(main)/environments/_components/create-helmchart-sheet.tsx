'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import dynamic from 'next/dynamic'
import { Package2, Loader2 } from 'lucide-react'
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
import { createHelmChart } from '@/app/actions/helmchart.actions'
import { toast } from 'sonner'

const CodeMirror = dynamic(() => import('@uiw/react-codemirror'), {
  ssr: false,
})

interface CreateHelmChartSheetProps {
  namespace: string
  user: string
}

const defaultValuesContent = `# Helm values (optional)
# Add your custom values here
`

export function CreateHelmChartSheet({ namespace, user }: CreateHelmChartSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [name, setName] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [description, setDescription] = useState('')
  const [chartName, setChartName] = useState('')
  const [chartUrl, setChartUrl] = useState('')
  const [chartVersion, setChartVersion] = useState('')
  const [valuesContent, setValuesContent] = useState(defaultValuesContent)
  const [yamlExtension, setYamlExtension] = useState<any>(null)

  useEffect(() => {
    import('@codemirror/lang-yaml').then((mod) => {
      setYamlExtension(mod.yaml())
    })
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      toast.error('Please enter a name')
      return
    }

    if (!displayName.trim()) {
      toast.error('Please enter a display name')
      return
    }

    if (!chartUrl.trim()) {
      toast.error('Please enter a chart repository URL')
      return
    }

    if (!chartName.trim()) {
      toast.error('Please enter a chart name')
      return
    }

    startTransition(async () => {
      const spec: any = {
        displayName: displayName.trim(),
        chart: {
          url: chartUrl.trim(),
          name: chartName.trim(),
        },
      }

      if (description.trim()) spec.description = description.trim()
      if (chartVersion.trim()) spec.chart.version = chartVersion.trim()

      // Parse YAML values to JSON if provided
      if (valuesContent.trim() && valuesContent.trim() !== defaultValuesContent.trim()) {
        try {
          const yaml = await import('js-yaml')
          spec.helmValues = yaml.load(valuesContent.trim())
        } catch (error) {
          toast.error('Invalid YAML in values')
          return
        }
      }

      const result = await createHelmChart(
        namespace,
        {
          name: name.trim(),
          spec,
        },
        user
      )

      if (result.success) {
        toast.success('Helm chart created successfully')
        setOpen(false)
        setName('')
        setDisplayName('')
        setDescription('')
        setChartName('')
        setChartUrl('')
        setChartVersion('')
        setValuesContent(defaultValuesContent)

        // Refresh and poll for updates
        router.refresh()

        let pollCount = 0
        const pollInterval = setInterval(() => {
          router.refresh()
          pollCount++
          if (pollCount >= 10) {
            clearInterval(pollInterval)
          }
        }, 1000)
      } else {
        toast.error(result.error || 'Failed to create helm chart')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button size="sm" className="gap-2">
          <Package2 className="h-4 w-4" />
          Add Chart
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-full sm:max-w-2xl">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Add Helm Chart</SheetTitle>
            <SheetDescription>
              Deploy a Helm chart from a repository
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-4 overflow-y-auto p-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name *</Label>
              <Input
                id="name"
                placeholder="my-chart"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isPending}
                required
              />
              <p className="text-xs text-muted-foreground">Kubernetes resource name (lowercase, alphanumeric, hyphens)</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="displayName">Display Name *</Label>
              <Input
                id="displayName"
                placeholder="My Helm Chart"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                disabled={isPending}
                required
              />
              <p className="text-xs text-muted-foreground">Human-readable name for the chart</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                placeholder="Description of the helm chart"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={isPending}
              />
              <p className="text-xs text-muted-foreground">Optional description</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="chartUrl">Chart Repository URL *</Label>
              <Input
                id="chartUrl"
                placeholder="https://charts.bitnami.com/bitnami"
                value={chartUrl}
                onChange={(e) => setChartUrl(e.target.value)}
                disabled={isPending}
                required
              />
              <p className="text-xs text-muted-foreground">Helm repository URL</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="chartName">Chart Name *</Label>
              <Input
                id="chartName"
                placeholder="nginx"
                value={chartName}
                onChange={(e) => setChartName(e.target.value)}
                disabled={isPending}
                required
              />
              <p className="text-xs text-muted-foreground">Chart name from the repository</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="chartVersion">Chart Version</Label>
              <Input
                id="chartVersion"
                placeholder="1.0.0"
                value={chartVersion}
                onChange={(e) => setChartVersion(e.target.value)}
                disabled={isPending}
              />
              <p className="text-xs text-muted-foreground">Chart version (optional, defaults to latest)</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="values-content">Values (YAML)</Label>
              <div className="rounded-md border">
                {yamlExtension ? (
                  <CodeMirror
                    value={valuesContent}
                    height="200px"
                    extensions={[yamlExtension]}
                    onChange={(value) => setValuesContent(value)}
                    className="text-sm"
                  />
                ) : (
                  <div className="h-[200px] flex items-center justify-center text-muted-foreground">
                    Loading editor...
                  </div>
                )}
              </div>
              <p className="text-xs text-muted-foreground">Custom values for the chart (optional)</p>
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
              Add Chart
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
