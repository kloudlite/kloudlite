import { Activity, Globe, GitBranch, Clock, Cpu, MemoryStick, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface Service {
  id: string
  name: string
  image: string
  replicas: number
  status: 'running' | 'stopped' | 'error'
  port: number
  cpu: string
  memory: string
  lastDeployed: string
}

const services: Service[] = [
  {
    id: '1',
    name: 'web-frontend',
    image: 'nginx:1.24-alpine',
    replicas: 3,
    status: 'running',
    port: 80,
    cpu: '100m',
    memory: '128Mi',
    lastDeployed: '2 hours ago'
  },
  {
    id: '2',
    name: 'api-backend',
    image: 'node:18-alpine',
    replicas: 2,
    status: 'running',
    port: 3000,
    cpu: '200m',
    memory: '256Mi',
    lastDeployed: '1 day ago'
  },
  {
    id: '3',
    name: 'postgres-db',
    image: 'postgres:15',
    replicas: 1,
    status: 'running',
    port: 5432,
    cpu: '500m',
    memory: '1Gi',
    lastDeployed: '1 week ago'
  },
  {
    id: '4',
    name: 'redis-cache',
    image: 'redis:7-alpine',
    replicas: 1,
    status: 'running',
    port: 6379,
    cpu: '100m',
    memory: '128Mi',
    lastDeployed: '3 days ago'
  },
]

export function EnvironmentServices() {
  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">Services</h3>
          <p className="text-sm text-gray-500 mt-1">Running services and applications in this environment</p>
        </div>
        <Button size="sm" className="gap-2">
          <Plus className="h-4 w-4" />
          Deploy Service
        </Button>
      </div>

      <div className="bg-white rounded-lg border border-gray-200">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 bg-gray-50">
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Service
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Image
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Replicas
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Port
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Resources
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Last Deployed
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {services.map((service) => (
                <tr key={service.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <Activity className="h-5 w-5 text-gray-400 mr-3" />
                      <span className="text-sm font-medium text-gray-900">{service.name}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-1">
                      <GitBranch className="h-3 w-3 text-gray-400" />
                      <span className="text-sm text-gray-600">{service.image}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      service.status === 'running'
                        ? 'bg-green-100 text-green-800'
                        : service.status === 'stopped'
                        ? 'bg-gray-100 text-gray-800'
                        : 'bg-red-100 text-red-800'
                    }`}>
                      {service.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="text-sm text-gray-600">{service.replicas}</span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-1 text-sm text-gray-600">
                      <Globe className="h-3 w-3 text-gray-400" />
                      {service.port}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-600">
                      <div className="flex items-center gap-2">
                        <div className="flex items-center gap-1">
                          <Cpu className="h-3 w-3 text-gray-400" />
                          {service.cpu}
                        </div>
                        <div className="flex items-center gap-1">
                          <MemoryStick className="h-3 w-3 text-gray-400" />
                          {service.memory}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-1 text-sm text-gray-500">
                      <Clock className="h-3 w-3" />
                      {service.lastDeployed}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}