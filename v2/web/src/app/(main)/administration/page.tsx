import { auth } from '@/lib/auth'
import { AdminWorkMachinesList } from '@/components/admin-work-machines-list'

// Mock data for work machines
const mockWorkMachines = [
  {
    id: '1',
    owner: 'john.doe@company.com',
    name: 'dev-machine-01',
    status: 'active' as const,
    cpu: 45,
    memory: 62,
    disk: 78,
    uptime: '2 days',
    type: 'standard',
    lastActive: new Date('2024-01-15T10:30:00'),
    workspaces: 3,
    environments: 2,
  },
  {
    id: '2',
    owner: 'sarah.smith@company.com',
    name: 'dev-machine-02',
    status: 'idle' as const,
    cpu: 12,
    memory: 28,
    disk: 45,
    uptime: '5 hours',
    type: 'performance',
    lastActive: new Date('2024-01-15T14:20:00'),
    workspaces: 5,
    environments: 3,
  },
  {
    id: '3',
    owner: 'mike.wilson@company.com',
    name: 'dev-machine-03',
    status: 'stopped' as const,
    cpu: 0,
    memory: 0,
    disk: 92,
    uptime: '0 minutes',
    type: 'basic',
    lastActive: new Date('2024-01-14T18:45:00'),
    workspaces: 1,
    environments: 1,
  },
  {
    id: '4',
    owner: 'emma.davis@company.com',
    name: 'dev-machine-04',
    status: 'active' as const,
    cpu: 78,
    memory: 85,
    disk: 65,
    uptime: '1 day',
    type: 'premium',
    lastActive: new Date('2024-01-15T15:00:00'),
    workspaces: 8,
    environments: 6,
  },
  {
    id: '5',
    owner: 'alex.johnson@company.com',
    name: 'dev-machine-05',
    status: 'idle' as const,
    cpu: 5,
    memory: 15,
    disk: 34,
    uptime: '3 hours',
    type: 'standard',
    lastActive: new Date('2024-01-15T12:30:00'),
    workspaces: 2,
    environments: 2,
  },
]

export default async function AdministrationPage() {
  const session = await auth()
  const isSuperAdmin = session?.user?.platformRole === 'super_admin'

  return (
    <main className="mx-auto max-w-7xl px-6 py-8">
      {/* Page Header */}
      <div className="mb-8">
        <div>
          <h1 className="text-3xl font-light tracking-tight">Administration</h1>
          <p className="text-sm text-gray-600 mt-2">
            Monitor and manage work machines across your organization
          </p>
        </div>
      </div>

      {/* Statistics Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <p className="text-sm text-gray-600">Total Machines</p>
          <p className="text-2xl font-semibold mt-1">{mockWorkMachines.length}</p>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <p className="text-sm text-gray-600">Active</p>
          <p className="text-2xl font-semibold mt-1 text-green-600">
            {mockWorkMachines.filter(m => m.status === 'active').length}
          </p>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <p className="text-sm text-gray-600">Idle</p>
          <p className="text-2xl font-semibold mt-1 text-yellow-600">
            {mockWorkMachines.filter(m => m.status === 'idle').length}
          </p>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <p className="text-sm text-gray-600">Stopped</p>
          <p className="text-2xl font-semibold mt-1 text-gray-600">
            {mockWorkMachines.filter(m => m.status === 'stopped').length}
          </p>
        </div>
      </div>

      {/* Work Machines List */}
      <AdminWorkMachinesList
        workMachines={mockWorkMachines}
        isSuperAdmin={isSuperAdmin}
      />
    </main>
  )
}