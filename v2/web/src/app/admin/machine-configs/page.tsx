import { MachineConfigsList } from '@/components/machine-configs-list'

// Mock data for machine configurations
const mockConfigs = [
  {
    id: '1',
    name: 'Basic',
    cpu: 2,
    memory: 4,
    storage: 50,
    maxInstances: 10,
    activeInstances: 5,
    pricePerHour: 0.10,
    description: 'Suitable for light development and testing'
  },
  {
    id: '2',
    name: 'Standard',
    cpu: 4,
    memory: 8,
    storage: 100,
    maxInstances: 8,
    activeInstances: 7,
    pricePerHour: 0.20,
    description: 'Recommended for most development workloads'
  },
  {
    id: '3',
    name: 'Performance',
    cpu: 8,
    memory: 16,
    storage: 200,
    maxInstances: 5,
    activeInstances: 2,
    pricePerHour: 0.40,
    description: 'For resource-intensive applications'
  },
  {
    id: '4',
    name: 'Premium',
    cpu: 16,
    memory: 32,
    storage: 500,
    maxInstances: 3,
    activeInstances: 1,
    pricePerHour: 0.80,
    description: 'Maximum performance for demanding workloads'
  },
]

export default function MachineConfigsPage() {
  return (
    <main className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-light tracking-tight">Machine Configurations</h1>
        <p className="text-sm text-gray-600 mt-2">
          Define machine types and resource allocations
        </p>
      </div>

      {/* Machine Configs List Component */}
      <MachineConfigsList configs={mockConfigs} />
    </main>
  )
}