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
        return 'text-success bg-success/10'
      case 'idle':
        return 'text-warning bg-warning/10'
      case 'stopped':
        return 'text-muted-foreground bg-muted'
      default:
        return 'text-muted-foreground bg-muted'
    }
  }

  const getUsageIndicator = (value: number) => {
    if (value > 80) return 'text-destructive'
    if (value > 60) return 'text-warning'
    return 'text-success'
  }

  const handleAction = (action: string, machineId: string) => {
    console.log(`${action} machine ${machineId}`)
    // Implement actual actions here
  }

  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-border bg-card">
        <div className="border-b border-border px-6 py-4">
          <h2 className="text-lg font-medium">Work Machines</h2>
        </div>

        {/* Filters */}
        <div className="border-b border-border px-6 py-4">
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
              <tr className="border-b border-border">
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-muted-foreground uppercase">
                  Machine
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-muted-foreground uppercase">
                  Owner
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-muted-foreground uppercase">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-muted-foreground uppercase">
                  Resources
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-muted-foreground uppercase">
                  Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-muted-foreground uppercase">
                  Activity
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium tracking-wider text-muted-foreground uppercase">
                  Last Active
                </th>
                <th className="relative px-6 py-3">
                  <span className="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border bg-card">
              {filteredMachines.map((machine) => (
                <tr key={machine.id} className="hover:bg-muted">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link
                      href={`/administration/machines/${machine.id}`}
                      className="group flex items-center gap-3"
                    >
                      <Server className="h-5 w-5 text-muted-foreground group-hover:text-foreground" />
                      <span className="text-sm font-medium text-foreground group-hover:text-info">
                        {machine.name}
                      </span>
                    </Link>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="text-sm text-muted-foreground">{machine.owner}</span>
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
                        <span className="text-muted-foreground">CPU:</span>
                        <span className={`font-medium ${getUsageIndicator(machine.cpu)}`}>
                          {machine.cpu}%
                        </span>
                      </div>
                      <div className="flex items-center gap-1">
                        <span className="text-muted-foreground">Mem:</span>
                        <span className={`font-medium ${getUsageIndicator(machine.memory)}`}>
                          {machine.memory}%
                        </span>
                      </div>
                      <div className="flex items-center gap-1">
                        <span className="text-muted-foreground">Disk:</span>
                        <span className={`font-medium ${getUsageIndicator(machine.disk)}`}>
                          {machine.disk}%
                        </span>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="inline-flex items-center rounded bg-muted px-2 py-0.5 text-xs font-medium text-foreground">
                      {machine.type}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-xs text-muted-foreground">
                      <div>{machine.workspaces} workspaces</div>
                      <div>{machine.environments} environments</div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
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
                            className="text-destructive"
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
                              className="text-destructive"
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
            <Server className="mx-auto h-12 w-12 text-muted-foreground" />
            <p className="mt-2 text-sm text-muted-foreground">No work machines found</p>
          </div>
        )}
      </div>
    </div>
  )
}
