import { ObjectReference } from './workspace'

export interface PortMapping {
  servicePort: number
  workspacePort: number
  protocol: string
}

export interface ServiceInterceptSpec {
  workspaceRef: ObjectReference
  serviceRef: ObjectReference
  portMappings: PortMapping[]
  status: 'active' | 'inactive'
}

export interface Condition {
  type: string
  status: string
  lastTransitionTime: string
  reason: string
  message: string
}

export interface ServiceInterceptStatus {
  phase: 'Creating' | 'Active' | 'Inactive' | 'Failed'
  message: string
  originalServiceSelector?: Record<string, string>
  affectedPodNames?: string[]
  workspacePodIP?: string
  workspacePodName?: string
  interceptStartTime?: string
  interceptEndTime?: string
  conditions?: Condition[]
}

export interface ServiceIntercept {
  metadata: {
    name: string
    namespace: string
    labels?: Record<string, string>
    annotations?: Record<string, string>
    creationTimestamp?: string
  }
  spec: ServiceInterceptSpec
  status?: ServiceInterceptStatus
}

export interface ListServiceInterceptsResponse {
  serviceIntercepts: ServiceIntercept[]
  count: number
}
