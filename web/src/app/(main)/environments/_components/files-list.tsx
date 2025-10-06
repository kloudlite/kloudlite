'use client'

import { File, Upload } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { FileInfo } from '@/types/environment'
import { AddFileSheet } from './add-file-sheet'
import { FilesTable } from './files-table'

interface FilesListProps {
  environmentId: string
  files: FileInfo[]
}

export function FilesList({ environmentId, files }: FilesListProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">File Configs</h3>
          <p className="text-sm text-muted-foreground">Configuration files mounted to containers</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm">
            <Upload className="h-4 w-4 mr-2" />
            Upload File
          </Button>
          <AddFileSheet environmentId={environmentId} />
        </div>
      </div>

      {files.length === 0 ? (
        <div className="text-center py-12 bg-muted/50 rounded-lg border">
          <File className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No configuration files</p>
          <div className="mt-4">
            <AddFileSheet environmentId={environmentId} />
          </div>
        </div>
      ) : (
        <FilesTable files={files} environmentId={environmentId} />
      )}
    </div>
  )
}
