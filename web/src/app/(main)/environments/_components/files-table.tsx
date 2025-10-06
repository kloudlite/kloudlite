import { File } from 'lucide-react'
import type { FileInfo } from '@/types/environment'
import { FileActions } from './file-actions'

interface FilesTableProps {
  files: FileInfo[]
  environmentId: string
}

export function FilesTable({ files, environmentId }: FilesTableProps) {
  return (
    <div className="bg-card rounded-lg border overflow-hidden">
      <table className="min-w-full">
        <thead className="bg-muted/50 border-b">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">File Name</th>
            <th className="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody className="divide-y">
          {files.map((file) => (
            <tr key={file.name} className="hover:bg-muted/50">
              <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div className="flex items-center gap-2">
                  <File className="h-4 w-4 text-muted-foreground" />
                  {file.name}
                </div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-right text-sm space-x-1">
                <FileActions
                  file={file}
                  environmentId={environmentId}
                />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
