'use client'

import { useState } from 'react'
import { Lock, FileText, Key, File, Plus, Trash2, Eye, EyeOff, Upload, Download, Edit2 } from 'lucide-react'
import { Button } from '@/components/ui/button'

type ConfigType = 'configs' | 'secrets' | 'files'
type ConfigEntry = {
  id: string
  key: string
  value: string
  isSecure: boolean
  mountPath?: string
}

type FileEntry = {
  id: string
  name: string
  mountPath: string
  size: string
  lastModified: string
}

export function EnvironmentConfigs() {
  const [activeSection, setActiveSection] = useState<ConfigType>('configs')
  const [showSecrets, setShowSecrets] = useState<{ [key: string]: boolean }>({})

  // Mock data
  const configs: ConfigEntry[] = [
    { id: '1', key: 'DATABASE_URL', value: 'postgresql://localhost:5432/myapp', isSecure: false },
    { id: '2', key: 'REDIS_HOST', value: 'redis.example.com', isSecure: false },
    { id: '3', key: 'API_ENDPOINT', value: 'https://api.example.com/v1', isSecure: false },
    { id: '4', key: 'LOG_LEVEL', value: 'debug', isSecure: false },
    { id: '5', key: 'MAX_WORKERS', value: '10', isSecure: false },
    { id: '6', key: 'CACHE_TTL', value: '3600', isSecure: false },
    { id: '7', key: 'NODE_ENV', value: 'production', isSecure: false },
  ]

  const secrets: ConfigEntry[] = [
    { id: '8', key: 'JWT_SECRET', value: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...', isSecure: true },
    { id: '9', key: 'AWS_SECRET_ACCESS_KEY', value: 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY', isSecure: true },
    { id: '10', key: 'STRIPE_SECRET_KEY', value: 'sk_test_51H5gJKL8...', isSecure: true },
    { id: '11', key: 'GITHUB_TOKEN', value: 'ghp_xxxxxxxxxxxxxxxxxxxx', isSecure: true },
    { id: '12', key: 'DATABASE_PASSWORD', value: 'supersecretpassword123', isSecure: true },
  ]

  const files: FileEntry[] = [
    { id: '13', name: 'nginx.conf', mountPath: '/etc/nginx/nginx.conf', size: '2.4 KB', lastModified: '2 hours ago' },
    { id: '14', name: 'ssl-cert.pem', mountPath: '/etc/ssl/certs/server.pem', size: '1.8 KB', lastModified: '1 day ago' },
    { id: '15', name: 'app.properties', mountPath: '/app/config/app.properties', size: '856 B', lastModified: '3 days ago' },
    { id: '16', name: 'server.key', mountPath: '/etc/ssl/private/server.key', size: '1.6 KB', lastModified: '1 week ago' },
  ]

  const sections = [
    { id: 'configs' as const, label: 'Config Maps', icon: FileText, count: configs.length },
    { id: 'secrets' as const, label: 'Secrets', icon: Lock, count: secrets.length },
    { id: 'files' as const, label: 'File Configs', icon: File, count: files.length },
  ]

  const toggleSecretVisibility = (id: string) => {
    setShowSecrets(prev => ({ ...prev, [id]: !prev[id] }))
  }

  return (
    <div className="flex gap-6">
      {/* Left Navigation */}
      <div className="w-48 flex-shrink-0">
        <nav className="space-y-1">
          {sections.map((section) => {
            const Icon = section.icon
            return (
              <button
                key={section.id}
                onClick={() => setActiveSection(section.id)}
                className={`
                  w-full flex items-center gap-3 px-3 py-2 text-sm rounded-md transition-colors
                  ${activeSection === section.id
                    ? 'bg-gray-100 text-gray-900 font-medium'
                    : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                  }
                `}
              >
                <Icon className="h-4 w-4 flex-shrink-0" />
                {section.label}
                <span className="ml-auto text-xs text-gray-500">{section.count}</span>
              </button>
            )
          })}
        </nav>
      </div>

      {/* Content Area */}
      <div className="flex-1">
        {activeSection === 'configs' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-lg font-medium">Config Maps</h3>
                <p className="text-sm text-gray-500">Environment configuration variables</p>
              </div>
              <Button variant="outline" size="sm">
                <Upload className="h-4 w-4 mr-2" />
                Import
              </Button>
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
        )}

        {activeSection === 'secrets' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-lg font-medium">Secrets</h3>
                <p className="text-sm text-gray-500">Encrypted sensitive information</p>
              </div>
              <Button variant="outline" size="sm">
                <Upload className="h-4 w-4 mr-2" />
                Import
              </Button>
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
        )}

        {activeSection === 'files' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-lg font-medium">File Configs</h3>
                <p className="text-sm text-gray-500">Configuration files mounted to containers</p>
              </div>
              <Button variant="outline" size="sm">
                <Upload className="h-4 w-4 mr-2" />
                Upload File
              </Button>
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
        )}
      </div>
    </div>
  )
}