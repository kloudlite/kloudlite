// HelmChart types matching Kloudlite environments.kloudlite.io/v1 API structure

export interface HelmChartMetadata {
  name: string
  namespace: string
  uid?: string
  resourceVersion?: string
  generation?: number
  creationTimestamp?: string
  deletionTimestamp?: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  managedFields?: any[]
}

export interface HelmChartInfo {
  url: string
  name: string
  version?: string
}

export interface HelmJobVars {
  nodeSelector?: Record<string, string>
  tolerations?: any[]
  affinity?: any
  resources?: any
}

export interface HelmChartSpec {
  // DisplayName is the human-readable name
  displayName: string

  // Description provides additional information
  description?: string

  // Chart configuration
  chart: HelmChartInfo

  // Helm values as JSON
  helmValues?: any

  // Job configuration
  jobVars?: HelmJobVars

  // Lifecycle hooks
  preInstall?: string
  postInstall?: string
  preUninstall?: string
  postUninstall?: string
}

export interface HelmChartCondition {
  type: string
  status: 'True' | 'False' | 'Unknown'
  reason?: string
  message?: string
  lastTransitionTime?: string
}

export interface HelmChartStatus {
  // From reconciler.Status
  checks?: Record<string, any>
  checkList?: string[]
  isReady?: boolean
  lastReadyGeneration?: number
  lastReconcileTime?: string
  message?: {
    RawMessage?: string
  }

  // HelmChart-specific status
  state?: 'pending' | 'installing' | 'installed' | 'upgrading' | 'failed' | 'uninstalling' | 'deleting'
  releaseName?: string
  installedVersion?: string
  lastInstallTime?: string
  deployedResources?: string[]
  conditions?: HelmChartCondition[]
  observedGeneration?: number
}

export interface HelmChart {
  metadata: HelmChartMetadata
  spec: HelmChartSpec
  status?: HelmChartStatus
}

// API Request types
export interface HelmChartCreateRequest {
  name: string
  spec: HelmChartSpec
}

export interface HelmChartUpdateRequest {
  spec: HelmChartSpec
}

// API Response types
export interface HelmChartListResponse {
  helmCharts: HelmChart[]
  count: number
}

export interface HelmChartResponse {
  message: string
  helmChart: HelmChart
}

export interface HelmChartDeleteResponse {
  name: string
  namespace: string
  message: string
}

export interface HelmChartStatusResponse {
  name: string
  namespace: string
  state?: string
  message?: string
  releaseName?: string
  installedVersion?: string
}

// UI Helper types
export type HelmChartState = 'pending' | 'installing' | 'installed' | 'upgrading' | 'failed' | 'uninstalling' | 'deleting' | 'unknown'

export interface HelmChartListItem extends HelmChart {
  // Computed fields for UI display
  displayStatus?: HelmChartState
  displayMessage?: string
  lastUpdated?: string
}
