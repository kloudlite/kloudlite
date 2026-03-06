export interface CodeAnalysisFinding {
  severity: string
  category: string
  file: string
  line: number
  title: string
  description: string
  recommendation: string
}

export interface CodeAnalysisReport {
  version: string
  type: string
  workspace: string
  analyzedAt: string
  summary: {
    score: number
    criticalCount: number
    highCount: number
    mediumCount: number
    lowCount: number
  }
  findings: CodeAnalysisFinding[]
}

export interface CodeAnalysisResponse {
  security: CodeAnalysisReport | null
  quality: CodeAnalysisReport | null
  status: {
    watching: boolean
    inProgress: boolean
    pendingAnalysis: boolean
    lastAnalysis?: string
  }
}
