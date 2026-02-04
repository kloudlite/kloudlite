'use client'

import { useState } from 'react'
import { Button } from '@kloudlite/ui'
import { FileCode } from 'lucide-react'
import { CompositionEditor } from './composition-editor'

interface CompositionEditorTriggerProps {
  environmentName: string
  composeContent: string | null
}

export function CompositionEditorTrigger({
  environmentName,
  composeContent,
}: CompositionEditorTriggerProps) {
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button variant="outline" size="sm" onClick={() => setOpen(true)} className="gap-2">
        <FileCode className="h-4 w-4" />
        Composition
      </Button>
      <CompositionEditor
        environmentName={environmentName}
        composeContent={composeContent}
        open={open}
        onOpenChange={setOpen}
      />
    </>
  )
}
