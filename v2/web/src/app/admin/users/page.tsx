import { UserManagementList } from '@/components/user-management-list'

// Mock data for users
const mockUsers = [
  {
    id: '1',
    name: 'John Doe',
    email: 'john.doe@company.com',
    role: 'admin',
    status: 'active',
    lastLogin: '2 mins ago',
    created: '6 months ago',
    machineType: 'standard',
    machineQuota: 2,
    storageQuota: 100,
  },
  {
    id: '2',
    name: 'Sarah Smith',
    email: 'sarah.smith@company.com',
    role: 'user',
    status: 'active',
    lastLogin: '15 mins ago',
    created: '4 months ago',
    machineType: 'basic',
    machineQuota: 1,
    storageQuota: 50,
  },
  {
    id: '3',
    name: 'Mike Wilson',
    email: 'mike.wilson@company.com',
    role: 'user',
    status: 'suspended',
    lastLogin: '5 days ago',
    created: '3 months ago',
    machineType: 'basic',
    machineQuota: 1,
    storageQuota: 50,
  },
  {
    id: '4',
    name: 'Emma Davis',
    email: 'emma.davis@company.com',
    role: 'admin',
    status: 'active',
    lastLogin: '1 hour ago',
    created: '5 months ago',
    machineType: 'premium',
    machineQuota: 5,
    storageQuota: 500,
  },
  {
    id: '5',
    name: 'Alex Johnson',
    email: 'alex.johnson@company.com',
    role: 'user',
    status: 'inactive',
    lastLogin: '2 weeks ago',
    created: '2 months ago',
    machineType: 'standard',
    machineQuota: 2,
    storageQuota: 100,
  },
]

export default function UsersPage() {
  return (
    <main className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-light tracking-tight">User Management</h1>
        <p className="text-sm text-gray-600 mt-2">
          Manage user accounts, roles, and permissions
        </p>
      </div>

      {/* Users List Component */}
      <UserManagementList users={mockUsers} />
    </main>
  )
}