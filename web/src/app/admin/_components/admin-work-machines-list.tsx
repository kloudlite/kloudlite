'use client'

import { useState } from 'react'
import Link from 'next/link'
import {
  Server,
  MoreVertical,
  Play,
  Square,
  RefreshCw,
  AlertCircle,
  Activity,
  Clock,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface WorkMachine {
  id: string
  owner: string
  name: string
  status: 'active' | 'idle' | 'stopped'
  cpu: number
  memory: number
  disk: number
  uptime: string
  type: string
  lastActive: Date
  workspaces: number
  environments: number
}

interface AdminWorkMachinesListProps {
  workMachines: WorkMachine[]
  isSuperAdmin?: boolean
}

export function AdminWorkMachinesList({ workMachines, isSuperAdmin }: AdminWorkMachinesListProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [typeFilter, setTypeFilter] = useState<string>('all')

  const filteredMachines = workMachines.filter((machine) => {
    const matchesSearch =
      machine.owner.toLowerCase().includes(searchQuery.toLowerCase()) ||
      machine.name.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesStatus = statusFilter === 'all' || machine.status === statusFilter
    const matchesType = typeFilter === 'all' || machine.type === typeFilter

    return matchesSearch && matchesStatus && matchesType
  })

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'text-green-600 bg-green-50'
      case 'idle':
        return 'text-yellow-600 bg-yellow-50'
      case 'stopped':
        return 'text-gray-600 bg-gray-50'
      default:
        return 'text-gray-600 bg-gray-50'
    }
  }

  const getUsageIndicator = (value: number) => {
    if (value > 80) return 'text-red-500'
    if (value > 60) return 'text-yellow-500'
    return 'text-green-500'
  }

  const handleAction = (action: string, machineId: string) => {
    console.log(`${action} machine ${machineId}`)
    // Implement actual actions here
  }

  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-gray-200 bg-white">
        <div className="border-b border-gray-200 px-6 py-4">
          <h2 className="text-lg font-medium">Work Machines</h2>
        </div>

        {/* Filters */}
        <div className="border-b border-gray-200 px-6 py-4">
          <div className="flex flex-col gap-4 sm:flex-row">
            <Input
              placeholder="Search by owner or machine name..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="sm:max-w-sm"
            />
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="sm:w-[150px]">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="idle">Idle</SelectItem>
                <SelectItem value="stopped">Stopped</SelectItem>
              </SelectContent>
            </Select>
            <Select value={typeFilter} onValueChange={setTypeFilter}>
              <SelectTrigger className="sm:w-[150px]">
                <SelectValue placeholder="Type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Types</SelectItem>
                <SelectItem value="basic">Basic</SelectItem>
                <SelectItem value="standard">Standard</SelectItem>
                <SelectItem value="performance">Performance</SelectItem>
                <SelectItem value="premium">Premium</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Table */}
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">
                  Machine
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">
                  Owner
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">
                  Resources
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">
                  Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">
                  Activity
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">
                  Last Active
                </th>
                <th className="relative px-6 py-3">
                  <span className="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {filteredMachines.map((machine) => (
                <tr key={machine.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link
                      href={`/administration/machines/${machine.id}`}
                      className="group flex items-center gap-3"
                    >
                      <Server className="h-5 w-5 text-gray-400 group-hover:text-gray-600" />
                      <span className="text-sm font-medium text-gray-900 group-hover:text-blue-600">
                        {machine.name}
                      </span>
                    </Link>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="text-sm text-gray-600">{machine.owner}</span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${getStatusColor(machine.status)}`}
                    >
                      {machine.status}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-4 text-xs">
                      <div className="flex items-center gap-1">
                        <span className="text-gray-500">CPU:</span>
                        <span className={`font-medium ${getUsageIndicator(machine.cpu)}`}>
                          {machine.cpu}%
                        </span>
                      </div>
                      <div className="flex items-center gap-1">
                        <span className="text-gray-500">Mem:</span>
                        <span className={`font-medium ${getUsageIndicator(machine.memory)}`}>
                          {machine.memory}%
                        </span>
                      </div>
                      <div className="flex items-center gap-1">
                        <span className="text-gray-500">Disk:</span>
                        <span className={`font-medium ${getUsageIndicator(machine.disk)}`}>
                          {machine.disk}%
                        </span>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="inline-flex items-center rounded bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-800">
                      {machine.type}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-xs text-gray-600">
                      <div>{machine.workspaces} workspaces</div>
                      <div>{machine.environments} environments</div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-1 text-xs text-gray-500">
                      <Clock className="h-3 w-3" />
                      {new Date(machine.lastActive).toLocaleString()}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-right text-sm font-medium whitespace-nowrap">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem asChild>
                          <Link href={`/administration/machines/${machine.id}`}>
                            <Activity className="mr-2 h-4 w-4" />
                            View Details
                          </Link>
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        {machine.status === 'stopped' ? (
                          <DropdownMenuItem onClick={() => handleAction('start', machine.id)}>
                            <Play className="mr-2 h-4 w-4" />
                            Start Machine
                          </DropdownMenuItem>
                        ) : (
                          <DropdownMenuItem
                            onClick={() => handleAction('stop', machine.id)}
                            className="text-red-600"
                          >
                            <Square className="mr-2 h-4 w-4" />
                            Stop Machine
                          </DropdownMenuItem>
                        )}
                        <DropdownMenuItem onClick={() => handleAction('restart', machine.id)}>
                          <RefreshCw className="mr-2 h-4 w-4" />
                          Restart Machine
                        </DropdownMenuItem>
                        {isSuperAdmin && (
                          <>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              onClick={() => handleAction('terminate', machine.id)}
                              className="text-red-600"
                            >
                              <AlertCircle className="mr-2 h-4 w-4" />
                              Terminate Machine
                            </DropdownMenuItem>
                          </>
                        )}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {filteredMachines.length === 0 && (
          <div className="px-6 py-12 text-center">
            <Server className="mx-auto h-12 w-12 text-gray-400" />
            <p className="mt-2 text-sm text-gray-600">No work machines found</p>
          </div>
        )}
      </div>
    </div>
  )
}
