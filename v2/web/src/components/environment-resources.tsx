'use client'

import { useState } from 'react'
import { FileCode, Package2, Plus, ExternalLink, GitBranch, Clock } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface HelmChart {
  id: string
  name: string
  version: string
  repository: string
  status: 'deployed' | 'pending' | 'failed'
  updatedAt: string
  values: number
}

interface DockerComposition {
  id: string
  name: string
  services: string[]
  status: 'running' | 'stopped'
  updatedAt: string
  networks: number
  volumes: number
}

const helmCharts: HelmChart[] = [
  {
    id: '1',
    name: 'nginx-ingress',
    version: '4.7.1',
    repository: 'https://kubernetes.github.io/ingress-nginx',
    status: 'deployed',
    updatedAt: '2 hours ago',
    values: 12
  },
  {
    id: '2',
    name: 'postgresql',
    version: '12.5.8',
    repository: 'https://charts.bitnami.com/bitnami',
    status: 'deployed',
    updatedAt: '5 days ago',
    values: 35
  },
  {
    id: '3',
    name: 'redis',
    version: '17.11.3',
    repository: 'https://charts.bitnami.com/bitnami',
    status: 'deployed',
    updatedAt: '1 week ago',
    values: 18
  },
]

const dockerCompositions: DockerComposition[] = [
  {
    id: '1',
    name: 'development-stack',
    services: ['web', 'api', 'postgres', 'redis'],
    status: 'running',
    updatedAt: '10 mins ago',
    networks: 2,
    volumes: 4
  },
  {
    id: '2',
    name: 'monitoring-stack',
    services: ['prometheus', 'grafana', 'loki'],
    status: 'running',
    updatedAt: '1 hour ago',
    networks: 1,
    volumes: 3
  },
  {
    id: '3',
    name: 'testing-services',
    services: ['mockserver', 'mailhog'],
    status: 'stopped',
    updatedAt: '3 days ago',
    networks: 1,
    volumes: 1
  },
]

export function EnvironmentResources() {
  const [activeSection, setActiveSection] = useState<'helm' | 'docker'>('helm')
  return (
    <div className="flex gap-6">
      {/* Left Navigation */}
      <div className="w-48 flex-shrink-0">
        <nav className="space-y-1">
          <button
            onClick={() => setActiveSection('helm')}
            className={`w-full flex items-center gap-2 px-3 py-2 text-sm font-medium rounded-md transition-colors ${
              activeSection === 'helm'
                ? 'bg-gray-100 text-gray-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
            }`}
          >
            <Package2 className="h-4 w-4" />
            Helm Charts
            <span className="ml-auto text-xs text-gray-500">{helmCharts.length}</span>
          </button>
          <button
            onClick={() => setActiveSection('docker')}
            className={`w-full flex items-center gap-2 px-3 py-2 text-sm font-medium rounded-md transition-colors ${
              activeSection === 'docker'
                ? 'bg-gray-100 text-gray-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
            }`}
          >
            <FileCode className="h-4 w-4" />
            Compositions
            <span className="ml-auto text-xs text-gray-500">{dockerCompositions.length}</span>
          </button>
        </nav>
      </div>

      {/* Content Area */}
      <div className="flex-1">
        {activeSection === 'helm' && (
          <div>
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-lg font-medium">Helm Charts</h3>
                <p className="text-sm text-gray-500 mt-1">Kubernetes applications deployed via Helm</p>
              </div>
              <Button size="sm" className="gap-2">
                <Plus className="h-4 w-4" />
                Add Chart
              </Button>
            </div>

        <div className="bg-white rounded-lg border border-gray-200">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200 bg-gray-50">
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Chart
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Version
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Repository
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Updated
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Values
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {helmCharts.map((chart) => (
                  <tr key={chart.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <Package2 className="h-5 w-5 text-gray-400 mr-3" />
                        <span className="text-sm font-medium text-gray-900">{chart.name}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm text-gray-600">{chart.version}</span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        chart.status === 'deployed'
                          ? 'bg-green-100 text-green-800'
                          : chart.status === 'pending'
                          ? 'bg-yellow-100 text-yellow-800'
                          : 'bg-red-100 text-red-800'
                      }`}>
                        {chart.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <a href={chart.repository} className="text-sm text-blue-600 hover:text-blue-500 flex items-center gap-1">
                        <span className="truncate max-w-[200px]">{chart.repository}</span>
                        <ExternalLink className="h-3 w-3" />
                      </a>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-1 text-sm text-gray-500">
                        <Clock className="h-3 w-3" />
                        {chart.updatedAt}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm text-gray-600">{chart.values} configs</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
          </div>
        )}

        {activeSection === 'docker' && (
          <div>
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-lg font-medium">Compositions</h3>
                <p className="text-sm text-gray-500 mt-1">Container stacks managed with Docker Compose</p>
              </div>
              <Button size="sm" className="gap-2">
                <Plus className="h-4 w-4" />
                Add Composition
              </Button>
            </div>

        <div className="bg-white rounded-lg border border-gray-200">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200 bg-gray-50">
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Composition
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Services
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Networks
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Volumes
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Updated
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {dockerCompositions.map((composition) => (
                  <tr key={composition.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <FileCode className="h-5 w-5 text-gray-400 mr-3" />
                        <span className="text-sm font-medium text-gray-900">{composition.name}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {composition.services.map((service, idx) => (
                          <span
                            key={idx}
                            className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700"
                          >
                            {service}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        composition.status === 'running'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}>
                        {composition.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm text-gray-600">{composition.networks}</span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm text-gray-600">{composition.volumes}</span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-1 text-sm text-gray-500">
                        <Clock className="h-3 w-3" />
                        {composition.updatedAt}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
          </div>
        )}
      </div>
    </div>
  )
}