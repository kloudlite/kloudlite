import { File } from 'lucide-react'
import type { FileInfo } from '@kloudlite/types'
import { AddFileSheet } from './add-file-sheet'
import { FilesTable } from './files-table'

interface FilesListProps {
  environmentId: string
  files: FileInfo[]
}

export function FilesList({ environmentId, files }: FilesListProps) {
  return (
    <div className="space-y-4">
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium">File Configs</h3>
          <p className="text-muted-foreground text-sm">Configuration files mounted to containers</p>
        </div>
        {files.length > 0 && <AddFileSheet environmentId={environmentId} />}
      </div>

      {files.length === 0 ? (
        <div className="bg-muted/50 rounded-lg border py-12 text-center">
          <File className="text-muted-foreground mx-auto mb-4 h-12 w-12" />
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
