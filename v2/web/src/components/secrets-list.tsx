'use client'

import { useState } from 'react'
import { Lock, Upload, Edit2, Trash2, Eye, EyeOff, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'

type SecretEntry = {
  id: string
  key: string
  value: string
}

const secrets: SecretEntry[] = [
  { id: '8', key: 'JWT_SECRET', value: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' },
  { id: '9', key: 'AWS_SECRET_ACCESS_KEY', value: 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY' },
  { id: '10', key: 'STRIPE_SECRET_KEY', value: 'sk_test_51H5gJKL8...' },
  { id: '11', key: 'GITHUB_TOKEN', value: 'ghp_xxxxxxxxxxxxxxxxxxxx' },
  { id: '12', key: 'DATABASE_PASSWORD', value: 'supersecretpassword123' },
]

export function SecretsList() {
  const [showSecrets, setShowSecrets] = useState<{ [key: string]: boolean }>({})

  const toggleSecretVisibility = (id: string) => {
    setShowSecrets(prev => ({ ...prev, [id]: !prev[id] }))
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">Secrets</h3>
          <p className="text-sm text-gray-500">Encrypted sensitive information</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm">
            <Upload className="h-4 w-4 mr-2" />
            Import
          </Button>
          <Button size="sm">
            <Plus className="h-4 w-4 mr-2" />
            Add Secret
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
            {secrets.map((secret) => (
              <tr key={secret.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                  <div className="flex items-center gap-2">
                    <Lock className="h-3 w-3 text-amber-500" />
                    {secret.key}
                  </div>
                </td>
                <td className="px-6 py-4 text-sm text-gray-600 font-mono">
                  <div className="flex items-center gap-2">
                    {showSecrets[secret.id] ? (
                      <span className="text-xs max-w-md truncate">{secret.value}</span>
                    ) : (
                      <span>••••••••••••••••</span>
                    )}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => toggleSecretVisibility(secret.id)}
                    >
                      {showSecrets[secret.id] ? (
                        <EyeOff className="h-3 w-3" />
                      ) : (
                        <Eye className="h-3 w-3" />
                      )}
                    </Button>
                  </div>
                </td>
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
