'use client'

import { useState, useRef } from 'react'
import { useRouter } from 'next/navigation'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@kloudlite/ui'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import { importEnvironmentConfig } from '@/app/actions/environment.actions'
import { toast } from 'sonner'
import { Upload, FileJson, AlertCircle } from 'lucide-react'

interface ImportEnvironmentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
  currentUser: string
}

interface ExportData {
  apiVersion?: string
  kind?: string
  metadata?: {
    name?: string
    exportedAt?: string
  }
  configs?: Record<string, string>
  secrets?: Record<string, string>
  files?: Array<{ name: string; content: string }>
  compositions?: Array<{ name: string; spec: unknown }>
}

export function ImportEnvironmentDialog({
  open,
  onOpenChange,
  onSuccess,
  currentUser,
}: ImportEnvironmentDialogProps) {
  const router = useRouter()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [newEnvName, setNewEnvName] = useState('')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [parsedData, setParsedData] = useState<ExportData | null>(null)
  const [parseError, setParseError] = useState<string | null>(null)

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setSelectedFile(file)
    setParseError(null)
    setParsedData(null)

    try {
      const content = await file.text()
      const data = JSON.parse(content) as ExportData

      // Validate the structure
      if (data.kind !== 'EnvironmentExport') {
        setParseError('Invalid file format. Expected an environment export file.')
        return
      }

      setParsedData(data)

      // Suggest a name based on the original
      if (data.metadata?.name) {
        const suggestedName = `${data.metadata.name}-copy`
        setNewEnvName(suggestedName)
      }
    } catch {
      setParseError('Failed to parse file. Please ensure it is a valid JSON export.')
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!parsedData || !newEnvName.trim()) return

    setIsSubmitting(true)

    try {
      // Don't pass targetNamespace - let the webhook auto-generate it
      const result = await importEnvironmentConfig(
        newEnvName.trim(),
        '', // Webhook will generate: env-{owner}--{name}
        currentUser,
        {
          configs: parsedData.configs,
          secrets: parsedData.secrets,
          files: parsedData.files,
          compositions: parsedData.compositions,
        },
      )

      if (result.success) {
        const warnings = 'warnings' in result ? result.warnings : undefined
        if (warnings && warnings.length > 0) {
          toast.success('Environment imported with some warnings', {
            description: `${newEnvName} was created but some items failed to import.`,
          })
        } else {
          toast.success('Environment imported', {
            description: `${newEnvName} has been created successfully.`,
          })
        }

        onOpenChange(false)
        resetForm()
        onSuccess?.()
        router.refresh()
      } else {
        toast.error('Failed to import environment', {
          description: result.error || 'An error occurred during import',
        })
      }
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Unknown error')
      toast.error('Failed to import environment', {
        description: error.message,
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  const resetForm = () => {
    setNewEnvName('')
    setSelectedFile(null)
    setParsedData(null)
    setParseError(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const handleOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      resetForm()
    }
    onOpenChange(isOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Import Environment</DialogTitle>
            <DialogDescription>
              Create a new environment from an exported configuration file.
            </DialogDescription>
          </DialogHeader>

          <div className="py-4 space-y-4">
            {/* File upload */}
            <div className="space-y-2">
              <Label>Configuration File</Label>
              <div
                onClick={() => fileInputRef.current?.click()}
                className={`border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors ${
                  parseError
                    ? 'border-destructive/50 bg-destructive/5'
                    : selectedFile
                      ? 'border-primary/50 bg-primary/5'
                      : 'border-muted-foreground/25 hover:border-muted-foreground/50'
                }`}
              >
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".json"
                  onChange={handleFileSelect}
                  className="hidden"
                />
                {selectedFile ? (
                  <div className="flex items-center justify-center gap-2">
                    <FileJson className="h-5 w-5 text-primary" />
                    <span className="text-sm font-medium">{selectedFile.name}</span>
                  </div>
                ) : (
                  <div className="flex flex-col items-center gap-2">
                    <Upload className="h-8 w-8 text-muted-foreground" />
                    <span className="text-sm text-muted-foreground">
                      Click to select a configuration file
                    </span>
                    <span className="text-xs text-muted-foreground/70">
                      JSON files exported from an environment
                    </span>
                  </div>
                )}
              </div>

              {parseError && (
                <div className="flex items-center gap-2 text-sm text-destructive">
                  <AlertCircle className="h-4 w-4" />
                  <span>{parseError}</span>
                </div>
              )}
            </div>

            {/* Parsed data preview */}
            {parsedData && (
              <div className="space-y-3 rounded-lg border bg-muted/50 p-4">
                <div className="text-sm font-medium">Import Summary</div>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div className="text-muted-foreground">Original Name:</div>
                  <div>{parsedData.metadata?.name || 'Unknown'}</div>

                  <div className="text-muted-foreground">Configs:</div>
                  <div>{Object.keys(parsedData.configs || {}).length} items</div>

                  <div className="text-muted-foreground">Secrets:</div>
                  <div>{Object.keys(parsedData.secrets || {}).length} items</div>

                  <div className="text-muted-foreground">Files:</div>
                  <div>{parsedData.files?.length || 0} items</div>

                  <div className="text-muted-foreground">Compositions:</div>
                  <div>{parsedData.compositions?.length || 0} items</div>
                </div>
              </div>
            )}

            {/* New environment name */}
            {parsedData && (
              <div className="space-y-2">
                <Label htmlFor="newEnvName">New Environment Name</Label>
                <Input
                  id="newEnvName"
                  value={newEnvName}
                  onChange={(e) => setNewEnvName(e.target.value)}
                  placeholder="Enter environment name"
                  disabled={isSubmitting}
                />
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting || !parsedData || !newEnvName.trim()}
            >
              {isSubmitting ? 'Importing...' : 'Import Environment'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
