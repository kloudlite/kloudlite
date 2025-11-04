// Work Machine types based on backend API
export interface WorkMachine {
  metadata: {
    name: string
    creationTimestamp?: string
    resourceVersion?: string
    uid?: string
    labels?: Record<string, string>
  }
  spec: WorkMachineSpec
  status?: WorkMachineStatus
}

export interface WorkMachineSpec {
  ownedBy: string
  machineType: string
  targetNamespace: string
  state: MachineState
  sshPublicKeys?: string[]
}

export type MachineState = 'starting' | 'running' | 'stopping' | 'stopped' | 'disabled' | 'errored'

export interface WorkMachineStatus {
  state?: MachineState
  conditions?: Array<{
    type: string
    status: string
    lastTransitionTime?: string
    reason?: string
    message?: string
  }>
  startedAt?: string
  stoppedAt?: string
  uptime?: string
  podName?: string
  podIP?: string
  sshPublicKey?: string
}

// API Response types
export interface WorkMachineListResponse {
  items: WorkMachine[]
  count: number
}

export interface WorkMachineResponse {
  success: boolean
  data?: WorkMachine
  error?: string
}
