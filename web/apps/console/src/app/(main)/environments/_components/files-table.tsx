import { File } from 'lucide-react'
import type { FileInfo } from '@kloudlite/types'
import { FileActions } from './file-actions'

interface FilesTableProps {
  files: FileInfo[]
  environmentId: string
}

export function FilesTable({ files, environmentId }: FilesTableProps) {
  return (
    <div className="bg-card overflow-hidden rounded-lg border">
      <table className="min-w-full">
        <thead className="bg-muted/50 border-b">
          <tr>
            <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
              File Name
            </th>
            <th className="text-muted-foreground px-6 py-3 text-right text-xs font-medium tracking-wider uppercase">
              Actions
            </th>
          </tr>
        </thead>
        <tbody className="divide-y">
          {files.map((file) => (
            <tr key={file.name} className="hover:bg-muted/50">
              <td className="px-6 py-4 text-sm font-medium whitespace-nowrap">
                <div className="flex items-center gap-2">
                  <File className="text-muted-foreground h-4 w-4" />
                  {file.name}
                </div>
              </td>
              <td className="space-x-1 px-6 py-4 text-right text-sm whitespace-nowrap">
                <FileActions file={file} environmentId={environmentId} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
