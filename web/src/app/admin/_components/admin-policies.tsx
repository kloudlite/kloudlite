'use client'

import { useState } from 'react'
import {
  Save,
  AlertCircle
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import {
  Alert,
  AlertDescription,
} from '@/components/ui/alert'

export function AdminPolicies() {
  const [machineTypes, setMachineTypes] = useState({
    basic: { cpu: 2, memory: 4, disk: 50, enabled: true },
    standard: { cpu: 4, memory: 8, disk: 100, enabled: true },
    performance: { cpu: 8, memory: 16, disk: 200, enabled: true },
    premium: { cpu: 16, memory: 32, disk: 500, enabled: true },
  })

  const [userPolicies, setUserPolicies] = useState({
    defaultMachineType: 'basic',
    maxMachinesPerUser: 1,
    autoStopIdleMinutes: 30,
    maxIdleMinutes: 120,
    allowTypeUpgrade: true,
    requireApprovalForPremium: true,
  })

  const [resourceQuotas, setResourceQuotas] = useState({
    maxWorkspacesPerMachine: 10,
    maxEnvironmentsPerMachine: 5,
    maxDiskUsageGB: 500,
    maxMemoryUsageGB: 32,
    maxCPUCores: 16,
  })

  const [isSaving, setIsSaving] = useState(false)

  const handleSave = async () => {
    setIsSaving(true)
    // Simulate save
    await new Promise(resolve => setTimeout(resolve, 1500))
    setIsSaving(false)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">Machine Policies</h1>
          <p className="text-sm text-gray-600 mt-1.5">
            Configure machine types, resource limits, and user policies
          </p>
        </div>
        <Button onClick={handleSave} disabled={isSaving} className="gap-2">
          <Save className="h-4 w-4" />
          {isSaving ? 'Saving...' : 'Save Changes'}
        </Button>
      </div>

      <Alert>
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Changes to policies will affect all users immediately. Existing machines will be updated on their next restart.
        </AlertDescription>
      </Alert>

      <Tabs defaultValue="machine-types" className="space-y-4">
        <TabsList>
          <TabsTrigger value="machine-types">Machine Types</TabsTrigger>
          <TabsTrigger value="user-policies">User Policies</TabsTrigger>
          <TabsTrigger value="resource-quotas">Resource Quotas</TabsTrigger>
          <TabsTrigger value="auto-scaling">Auto Scaling</TabsTrigger>
        </TabsList>

        <TabsContent value="machine-types" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Available Machine Types</CardTitle>
              <CardDescription>
                Configure the specifications and availability of each machine type
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {Object.entries(machineTypes).map(([type, specs]) => (
                <div key={type} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-medium capitalize">{type}</h3>
                    <Switch
                      checked={specs.enabled}
                      onCheckedChange={(checked) =>
                        setMachineTypes(prev => ({
                          ...prev,
                          [type]: { ...prev[type as keyof typeof prev], enabled: checked }
                        }))
                      }
                    />
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <Label htmlFor={`${type}-cpu`}>CPU Cores</Label>
                      <Input
                        id={`${type}-cpu`}
                        type="number"
                        value={specs.cpu}
                        onChange={(e) =>
                          setMachineTypes(prev => ({
                            ...prev,
                            [type]: { ...prev[type as keyof typeof prev], cpu: parseInt(e.target.value) }
                          }))
                        }
                        disabled={!specs.enabled}
                      />
                    </div>
                    <div>
                      <Label htmlFor={`${type}-memory`}>Memory (GB)</Label>
                      <Input
                        id={`${type}-memory`}
                        type="number"
                        value={specs.memory}
                        onChange={(e) =>
                          setMachineTypes(prev => ({
                            ...prev,
                            [type]: { ...prev[type as keyof typeof prev], memory: parseInt(e.target.value) }
                          }))
                        }
                        disabled={!specs.enabled}
                      />
                    </div>
                    <div>
                      <Label htmlFor={`${type}-disk`}>Disk (GB)</Label>
                      <Input
                        id={`${type}-disk`}
                        type="number"
                        value={specs.disk}
                        onChange={(e) =>
                          setMachineTypes(prev => ({
                            ...prev,
                            [type]: { ...prev[type as keyof typeof prev], disk: parseInt(e.target.value) }
                          }))
                        }
                        disabled={!specs.enabled}
                      />
                    </div>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="user-policies" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>User Policies</CardTitle>
              <CardDescription>
                Configure default settings and restrictions for users
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <Label htmlFor="default-type">Default Machine Type</Label>
                  <Select
                    value={userPolicies.defaultMachineType}
                    onValueChange={(value) =>
                      setUserPolicies(prev => ({ ...prev, defaultMachineType: value }))
                    }
                  >
                    <SelectTrigger id="default-type">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="basic">Basic</SelectItem>
                      <SelectItem value="standard">Standard</SelectItem>
                      <SelectItem value="performance">Performance</SelectItem>
                      <SelectItem value="premium">Premium</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="max-machines">Max Machines per User</Label>
                  <Input
                    id="max-machines"
                    type="number"
                    value={userPolicies.maxMachinesPerUser}
                    onChange={(e) =>
                      setUserPolicies(prev => ({ ...prev, maxMachinesPerUser: parseInt(e.target.value) }))
                    }
                  />
                </div>
              </div>

              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <Label>Allow Type Upgrade</Label>
                    <p className="text-sm text-gray-500">Users can upgrade their machine type</p>
                  </div>
                  <Switch
                    checked={userPolicies.allowTypeUpgrade}
                    onCheckedChange={(checked) =>
                      setUserPolicies(prev => ({ ...prev, allowTypeUpgrade: checked }))
                    }
                  />
                </div>
                <div className="flex items-center justify-between">
                  <div>
                    <Label>Require Approval for Premium</Label>
                    <p className="text-sm text-gray-500">Premium machines need admin approval</p>
                  </div>
                  <Switch
                    checked={userPolicies.requireApprovalForPremium}
                    onCheckedChange={(checked) =>
                      setUserPolicies(prev => ({ ...prev, requireApprovalForPremium: checked }))
                    }
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <Label htmlFor="auto-stop">Auto-stop Idle (minutes)</Label>
                  <Input
                    id="auto-stop"
                    type="number"
                    value={userPolicies.autoStopIdleMinutes}
                    onChange={(e) =>
                      setUserPolicies(prev => ({ ...prev, autoStopIdleMinutes: parseInt(e.target.value) }))
                    }
                  />
                  <p className="text-sm text-gray-500 mt-1">
                    Machines will auto-stop after this idle time
                  </p>
                </div>
                <div>
                  <Label htmlFor="max-idle">Max Idle Time (minutes)</Label>
                  <Input
                    id="max-idle"
                    type="number"
                    value={userPolicies.maxIdleMinutes}
                    onChange={(e) =>
                      setUserPolicies(prev => ({ ...prev, maxIdleMinutes: parseInt(e.target.value) }))
                    }
                  />
                  <p className="text-sm text-gray-500 mt-1">
                    Force stop machines after this idle time
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="resource-quotas" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Resource Quotas</CardTitle>
              <CardDescription>
                Set maximum resource limits per machine
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <Label htmlFor="max-workspaces">Max Workspaces per Machine</Label>
                  <Input
                    id="max-workspaces"
                    type="number"
                    value={resourceQuotas.maxWorkspacesPerMachine}
                    onChange={(e) =>
                      setResourceQuotas(prev => ({ ...prev, maxWorkspacesPerMachine: parseInt(e.target.value) }))
                    }
                  />
                </div>
                <div>
                  <Label htmlFor="max-environments">Max Environments per Machine</Label>
                  <Input
                    id="max-environments"
                    type="number"
                    value={resourceQuotas.maxEnvironmentsPerMachine}
                    onChange={(e) =>
                      setResourceQuotas(prev => ({ ...prev, maxEnvironmentsPerMachine: parseInt(e.target.value) }))
                    }
                  />
                </div>
                <div>
                  <Label htmlFor="max-disk">Max Disk Usage (GB)</Label>
                  <Input
                    id="max-disk"
                    type="number"
                    value={resourceQuotas.maxDiskUsageGB}
                    onChange={(e) =>
                      setResourceQuotas(prev => ({ ...prev, maxDiskUsageGB: parseInt(e.target.value) }))
                    }
                  />
                </div>
                <div>
                  <Label htmlFor="max-memory">Max Memory Usage (GB)</Label>
                  <Input
                    id="max-memory"
                    type="number"
                    value={resourceQuotas.maxMemoryUsageGB}
                    onChange={(e) =>
                      setResourceQuotas(prev => ({ ...prev, maxMemoryUsageGB: parseInt(e.target.value) }))
                    }
                  />
                </div>
                <div>
                  <Label htmlFor="max-cpu">Max CPU Cores</Label>
                  <Input
                    id="max-cpu"
                    type="number"
                    value={resourceQuotas.maxCPUCores}
                    onChange={(e) =>
                      setResourceQuotas(prev => ({ ...prev, maxCPUCores: parseInt(e.target.value) }))
                    }
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="auto-scaling" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Auto Scaling</CardTitle>
              <CardDescription>
                Configure automatic resource scaling based on usage
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <Label>Enable Auto Scaling</Label>
                    <p className="text-sm text-gray-500">Automatically adjust resources based on usage</p>
                  </div>
                  <Switch />
                </div>
                <div className="flex items-center justify-between">
                  <div>
                    <Label>Scale Down When Idle</Label>
                    <p className="text-sm text-gray-500">Reduce resources for idle machines</p>
                  </div>
                  <Switch />
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <Label htmlFor="scale-up-threshold">Scale Up Threshold (%)</Label>
                  <Input
                    id="scale-up-threshold"
                    type="number"
                    defaultValue="80"
                  />
                  <p className="text-sm text-gray-500 mt-1">
                    Scale up when usage exceeds this threshold
                  </p>
                </div>
                <div>
                  <Label htmlFor="scale-down-threshold">Scale Down Threshold (%)</Label>
                  <Input
                    id="scale-down-threshold"
                    type="number"
                    defaultValue="20"
                  />
                  <p className="text-sm text-gray-500 mt-1">
                    Scale down when usage falls below this threshold
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}