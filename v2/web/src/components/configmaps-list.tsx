'use client'

import { Upload, Edit2, Trash2, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'

type ConfigEntry = {
  id: string
  key: string
  value: string
}

const configs: ConfigEntry[] = [
  { id: '1', key: 'DATABASE_URL', value: 'postgresql://localhost:5432/myapp' },
  { id: '2', key: 'REDIS_HOST', value: 'redis.example.com' },
  { id: '3', key: 'API_ENDPOINT', value: 'https://api.example.com/v1' },
  { id: '4', key: 'LOG_LEVEL', value: 'debug' },
  { id: '5', key: 'MAX_WORKERS', value: '10' },
  { id: '6', key: 'CACHE_TTL', value: '3600' },
  { id: '7', key: 'NODE_ENV', value: 'production' },
]

export function ConfigMapsList() {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">Config Maps</h3>
          <p className="text-sm text-gray-500">Environment configuration variables</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm">
            <Upload className="h-4 w-4 mr-2" />
            Import
          </Button>
          <Button size="sm">
            <Plus className="h-4 w-4 mr-2" />
            Add Config
          </Button>
        </div>
      </div>

      <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        <table className="min-w-full">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Key</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Value</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {configs.map((config) => (
              <tr key={config.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">{config.key}</td>
                <td className="px-6 py-4 text-sm text-gray-600 font-mono max-w-md truncate">{config.value}</td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                  <Button variant="ghost" size="sm" className="mr-2">
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
