export interface ResourceReference {
  name: string
  namespace?: string
}

export interface UserPreferencesSpec {
  pinnedWorkspaces?: ResourceReference[]
  pinnedEnvironments?: string[]
}

export interface UserPreferencesStatus {
  lastUpdated?: string
}

export interface UserPreferences {
  metadata: {
    name: string
    creationTimestamp?: string
    resourceVersion?: string
  }
  spec: UserPreferencesSpec
  status?: UserPreferencesStatus
}

export interface PinWorkspaceRequest {
  name: string
  namespace: string
}

export interface PinEnvironmentRequest {
  name: string
}
