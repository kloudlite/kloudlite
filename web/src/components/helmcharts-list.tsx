import { Package2, Plus, ExternalLink, Clock } from 'lucide-react'
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

export function HelmChartsList() {
  return (
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
  )
}
