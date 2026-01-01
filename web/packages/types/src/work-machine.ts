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

export interface AutoShutdownConfig {
  enabled: boolean
  idleThresholdMinutes: number
}

export interface WorkMachineSpec {
  ownedBy: string
  machineType: string
  targetNamespace: string
  state: MachineState
  sshPublicKeys?: string[]
  autoShutdown?: AutoShutdownConfig
  // Volume size for btrfs storage in GB (default: 100, min: 50, max: 1000)
  // Root volume is fixed at 50GB for OS only
  volumeSize?: number
  // Volume type (e.g., gp3, gp2, io1 for AWS)
  volumeType?: string
}

export type MachineState = 'starting' | 'running' | 'stopping' | 'stopped' | 'disabled' | 'errored'

export interface WorkMachineStatus {
  isReady?: boolean
  state?: MachineState
  conditions?: Array<{
    type: string
    status: string
    lastTransitionTime?: string
    reason?: string
    message?: string
  }>
  checks?: Record<string, {
    status: boolean
    message?: string
  }>
  startedAt?: string
  stoppedAt?: string
  uptime?: string
  podName?: string
  podIP?: string
  sshPublicKey?: string
  allIdleSince?: string
  isAutoStopped?: boolean
  // Storage volume size in GB (btrfs volume for PVCs and snapshots)
  storageVolumeSize?: number
  // Public IP of the instance
  publicIP?: string
  // Private IP of the instance
  privateIP?: string
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
