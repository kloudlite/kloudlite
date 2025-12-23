'use client'

import { useEffect, useState, useCallback } from 'react'
import { Shield, Code2, RefreshCw, AlertTriangle, CheckCircle2, XCircle, Loader2 } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import {
  getCodeAnalysis,
  triggerCodeAnalysis,
  type CodeAnalysisResponse,
  type CodeAnalysisReport,
} from '@/app/actions/workspace.actions'

interface CodeAnalysisCardProps {
  workspaceName: string
  namespace: string
}

function SeverityBadge({ severity }: { severity: string }) {
  const colors: Record<string, string> = {
    critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    high: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
    medium: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
    low: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  }

  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${colors[severity] || 'bg-secondary text-secondary-foreground'}`}
    >
      {severity}
    </span>
  )
}

function ScoreIndicator({ score }: { score: number }) {
  let color = 'text-green-600 dark:text-green-400'
  let bgColor = 'bg-green-100 dark:bg-green-900/30'
  let Icon = CheckCircle2

  if (score < 50) {
    color = 'text-red-600 dark:text-red-400'
    bgColor = 'bg-red-100 dark:bg-red-900/30'
    Icon = XCircle
  } else if (score < 80) {
    color = 'text-yellow-600 dark:text-yellow-400'
    bgColor = 'bg-yellow-100 dark:bg-yellow-900/30'
    Icon = AlertTriangle
  }

  return (
    <div className={`flex items-center gap-2 rounded-lg px-3 py-2 ${bgColor}`}>
      <Icon className={`h-5 w-5 ${color}`} />
      <span className={`text-lg font-semibold ${color}`}>{score}</span>
      <span className="text-muted-foreground text-xs">/100</span>
    </div>
  )
}

function ReportSummary({ report, type }: { report: CodeAnalysisReport; type: 'security' | 'quality' }) {
  const Icon = type === 'security' ? Shield : Code2
  const title = type === 'security' ? 'Security' : 'Code Quality'

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Icon className="text-primary h-4 w-4" />
          <span className="text-sm font-medium">{title}</span>
        </div>
        <ScoreIndicator score={report.summary.score} />
      </div>

      {/* Issue counts */}
      <div className="flex flex-wrap gap-2">
        {report.summary.criticalCount > 0 && (
          <div className="flex items-center gap-1">
            <SeverityBadge severity="critical" />
            <span className="text-muted-foreground text-xs">{report.summary.criticalCount}</span>
          </div>
        )}
        {report.summary.highCount > 0 && (
          <div className="flex items-center gap-1">
            <SeverityBadge severity="high" />
            <span className="text-muted-foreground text-xs">{report.summary.highCount}</span>
          </div>
        )}
        {report.summary.mediumCount > 0 && (
          <div className="flex items-center gap-1">
            <SeverityBadge severity="medium" />
            <span className="text-muted-foreground text-xs">{report.summary.mediumCount}</span>
          </div>
        )}
        {report.summary.lowCount > 0 && (
          <div className="flex items-center gap-1">
            <SeverityBadge severity="low" />
            <span className="text-muted-foreground text-xs">{report.summary.lowCount}</span>
          </div>
        )}
        {report.summary.criticalCount === 0 &&
          report.summary.highCount === 0 &&
          report.summary.mediumCount === 0 &&
          report.summary.lowCount === 0 && (
            <span className="text-muted-foreground text-xs">No issues found</span>
          )}
      </div>

      {/* Top findings */}
      {report.findings.length > 0 && (
        <div className="space-y-2">
          <p className="text-muted-foreground text-xs font-medium">Top Issues:</p>
          {report.findings.slice(0, 3).map((finding, idx) => (
            <div key={idx} className="rounded-md border bg-muted/30 p-2">
              <div className="flex items-start justify-between gap-2">
                <p className="text-xs font-medium">{finding.title}</p>
                <SeverityBadge severity={finding.severity} />
              </div>
              <p className="text-muted-foreground mt-1 text-xs">{finding.file}:{finding.line}</p>
            </div>
          ))}
          {report.findings.length > 3 && (
            <p className="text-muted-foreground text-xs">
              +{report.findings.length - 3} more issues
            </p>
          )}
        </div>
      )}

      {/* Last analyzed */}
      <p className="text-muted-foreground text-xs">
        Analyzed {new Date(report.analyzedAt).toLocaleString()}
      </p>
    </div>
  )
}

export function CodeAnalysisCard({ workspaceName, namespace }: CodeAnalysisCardProps) {
  const [data, setData] = useState<CodeAnalysisResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [triggering, setTriggering] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchAnalysis = useCallback(async () => {
    try {
      const result = await getCodeAnalysis(workspaceName, namespace)
      if (result.success && result.data) {
        setData(result.data)
        setError(null)
      } else {
        setError(result.error || 'Failed to load analysis')
      }
    } catch (err) {
      setError('Failed to load analysis')
    } finally {
      setLoading(false)
    }
  }, [workspaceName, namespace])

  useEffect(() => {
    fetchAnalysis()
    // Poll every 30 seconds if analysis is in progress
    const interval = setInterval(() => {
      if (data?.status?.inProgress || data?.status?.pendingAnalysis) {
        fetchAnalysis()
      }
    }, 30000)
    return () => clearInterval(interval)
  }, [fetchAnalysis, data?.status?.inProgress, data?.status?.pendingAnalysis])

  const handleTriggerAnalysis = async () => {
    setTriggering(true)
    try {
      const result = await triggerCodeAnalysis(workspaceName, namespace)
      if (result.success) {
        // Refresh data after triggering
        setTimeout(fetchAnalysis, 2000)
      }
    } finally {
      setTriggering(false)
    }
  }

  if (loading) {
    return (
      <div className="bg-card rounded-lg border p-6">
        <div className="flex items-center gap-2">
          <Loader2 className="text-muted-foreground h-4 w-4 animate-spin" />
          <span className="text-muted-foreground text-sm">Loading analysis...</span>
        </div>
      </div>
    )
  }

  const hasReports = data?.security || data?.quality
  const isAnalyzing = data?.status?.inProgress || data?.status?.pendingAnalysis

  return (
    <div className="bg-card rounded-lg border">
      <div className="border-b p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="bg-primary/10 rounded-lg p-2">
              <Shield className="text-primary h-4 w-4" />
            </div>
            <div>
              <h3 className="text-sm font-semibold">Code Analysis</h3>
              <p className="text-muted-foreground text-xs">
                {isAnalyzing
                  ? 'Analysis in progress...'
                  : hasReports
                    ? 'Security & quality reports'
                    : 'No analysis yet'}
              </p>
            </div>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={handleTriggerAnalysis}
            disabled={triggering || isAnalyzing}
          >
            {triggering || isAnalyzing ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="h-4 w-4" />
            )}
            <span className="ml-2">{isAnalyzing ? 'Analyzing...' : 'Analyze'}</span>
          </Button>
        </div>
      </div>

      <div className="space-y-4 p-4">
        {error && !hasReports && (
          <div className="text-muted-foreground text-center text-sm">
            <p>Code analysis not available</p>
            <p className="text-xs">Analysis will start automatically when files change</p>
          </div>
        )}

        {data?.security && <ReportSummary report={data.security} type="security" />}

        {data?.security && data?.quality && <div className="border-t" />}

        {data?.quality && <ReportSummary report={data.quality} type="quality" />}

        {!hasReports && !error && (
          <div className="text-muted-foreground py-4 text-center text-sm">
            <p>No analysis reports yet</p>
            <p className="text-xs">Click Analyze to run code analysis</p>
          </div>
        )}
      </div>
    </div>
  )
}
