import { notFound } from 'next/navigation'
import { AdminMachineDetail } from '@/components/admin-machine-detail'

// Mock data
const mockMachine = {
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
  workspaces: [
    {
      id: 'ws1',
      name: 'frontend-app',
      status: 'running',
      resources: { cpu: 15, memory: 20 },
      lastAccess: '10 minutes ago',
    },
    {
      id: 'ws2',
      name: 'backend-api',
      status: 'running',
      resources: { cpu: 20, memory: 30 },
      lastAccess: '2 hours ago',
    },
    {
      id: 'ws3',
      name: 'mobile-app',
      status: 'stopped',
      resources: { cpu: 0, memory: 0 },
      lastAccess: '1 day ago',
    },
  ],
  environments: [
    {
      id: 'env1',
      name: 'staging',
      status: 'running',
      resources: { cpu: 10, memory: 12 },
    },
    {
      id: 'env2',
      name: 'development',
      status: 'running',
      resources: { cpu: 0, memory: 0 },
    },
  ],
  activityLog: [
    { timestamp: new Date('2024-01-15T15:00:00'), action: 'Machine started', user: 'john.doe@company.com' },
    { timestamp: new Date('2024-01-15T14:45:00'), action: 'Type changed from basic to standard', user: 'admin@company.com' },
    { timestamp: new Date('2024-01-15T10:30:00'), action: 'Workspace "frontend-app" created', user: 'john.doe@company.com' },
    { timestamp: new Date('2024-01-14T18:00:00'), action: 'Machine stopped (auto-idle)', user: 'system' },
    { timestamp: new Date('2024-01-14T09:00:00'), action: 'Machine started', user: 'john.doe@company.com' },
  ],
}

interface PageProps {
  params: {
    id: string
  }
}

export default async function AdminMachineDetailPage({ params }: PageProps) {
  // In real app, fetch machine data by ID
  if (params.id !== '1') {
    notFound()
  }

  return (
    <main className="mx-auto max-w-7xl px-6 py-8">
      <AdminMachineDetail machine={mockMachine} />
    </main>
  )
}