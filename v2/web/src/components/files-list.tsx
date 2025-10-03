'use client'

import { File, Upload, Edit2, Trash2, Download, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'

type FileEntry = {
  id: string
  name: string
  mountPath: string
  size: string
  lastModified: string
}

const files: FileEntry[] = [
  { id: '13', name: 'nginx.conf', mountPath: '/etc/nginx/nginx.conf', size: '2.4 KB', lastModified: '2 hours ago' },
  { id: '14', name: 'ssl-cert.pem', mountPath: '/etc/ssl/certs/server.pem', size: '1.8 KB', lastModified: '1 day ago' },
  { id: '15', name: 'app.properties', mountPath: '/app/config/app.properties', size: '856 B', lastModified: '3 days ago' },
  { id: '16', name: 'server.key', mountPath: '/etc/ssl/private/server.key', size: '1.6 KB', lastModified: '1 week ago' },
]

export function FilesList() {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">File Configs</h3>
          <p className="text-sm text-gray-500">Configuration files mounted to containers</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm">
            <Upload className="h-4 w-4 mr-2" />
            Upload File
          </Button>
          <Button size="sm">
            <Plus className="h-4 w-4 mr-2" />
            Add File
          </Button>
        </div>
      </div>

      <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        <table className="min-w-full">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">File Name</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Mount Path</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Size</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Modified</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {files.map((file) => (
              <tr key={file.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  <div className="flex items-center gap-2">
                    <File className="h-4 w-4 text-gray-400" />
                    {file.name}
                  </div>
                </td>
                <td className="px-6 py-4 text-sm text-gray-600 font-mono text-xs">{file.mountPath}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{file.size}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{file.lastModified}</td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm space-x-1">
                  <Button variant="ghost" size="sm">
                    <Download className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="sm">
                    <Edit2 className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700">
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
