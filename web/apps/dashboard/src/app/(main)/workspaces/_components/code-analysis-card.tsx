'use client'

import { useEffect, useState, useCallback } from 'react'
import { Shield, Code2, RefreshCw, AlertTriangle, CheckCircle2, XCircle, Loader2, ArrowRight } from 'lucide-react'
import Link from 'next/link'
import { Button } from '@kloudlite/ui'
import {
  getCodeAnalysis,
  triggerCodeAnalysis,
  type CodeAnalysisResponse,
} from '@/app/actions/workspace.actions'

interface CodeAnalysisCardProps {
  workspaceName: string
  namespace: string
}

function SeverityBadge({ severity, count }: { severity: string; count: number }) {
  if (count === 0) return null

  const colors: Record<string, string> = {
    critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    high: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
    medium: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
    low: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  }

  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${colors[severity] || 'bg-secondary text-secondary-foreground'}`}
    >
      <span>{count}</span>
      <span>{severity}</span>
    </span>
  )
}

function ScoreBadge({ score }: { score: number }) {
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
    <div className={`flex items-center gap-1.5 rounded-lg px-2 py-1 ${bgColor}`}>
      <Icon className={`h-3.5 w-3.5 ${color}`} />
      <span className={`text-sm font-semibold ${color}`}>{score}</span>
      <span className="text-muted-foreground text-xs">/100</span>
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
    } catch (_err) {
      setError('Failed to load analysis')
    } finally {
      setLoading(false)
    }
  }, [workspaceName, namespace])

  useEffect(() => {
    fetchAnalysis()
    const interval = setInterval(() => {
      if (data?.status?.inProgress || data?.status?.pendingAnalysis) {
        fetchAnalysis()
      }
    }, 30000)
    return () => clearInterval(interval)
  }, [fetchAnalysis, data?.status?.inProgress, data?.status?.pendingAnalysis])

  const handleTriggerAnalysis = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setTriggering(true)
    try {
      const result = await triggerCodeAnalysis(workspaceName, namespace)
      if (result.success) {
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

  // Calculate totals
  const securityTotal =
    (data?.security?.summary?.criticalCount || 0) +
    (data?.security?.summary?.highCount || 0) +
    (data?.security?.summary?.mediumCount || 0) +
    (data?.security?.summary?.lowCount || 0)

  const qualityTotal =
    (data?.quality?.summary?.criticalCount || 0) +
    (data?.quality?.summary?.highCount || 0) +
    (data?.quality?.summary?.mediumCount || 0) +
    (data?.quality?.summary?.lowCount || 0)

  const totalIssues = securityTotal + qualityTotal

  const detailPageUrl = `/workspaces/${namespace}/${workspaceName}/code-analysis`

  return (
    <Link href={detailPageUrl} className="block">
      <div className="bg-card rounded-lg border transition-colors hover:border-primary/50 hover:bg-accent/50">
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
                      ? `${totalIssues} issues found`
                      : 'No analysis yet'}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              {hasReports && data?.quality?.summary?.score !== undefined && (
                <ScoreBadge score={data.quality.summary.score} />
              )}
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
              </Button>
            </div>
          </div>
        </div>

        <div className="p-4">
          {error && !hasReports ? (
            <div className="text-muted-foreground text-center text-sm">
              <p>Code analysis not available</p>
              <p className="text-xs">Click refresh to run analysis</p>
            </div>
          ) : hasReports ? (
            <div className="space-y-3">
              {/* Security Summary */}
              {data?.security && (
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Shield className="h-4 w-4 text-red-500" />
                    <span className="text-sm font-medium">Security</span>
                    <span className="text-muted-foreground text-xs">({securityTotal} issues)</span>
                  </div>
                  <div className="flex flex-wrap gap-1">
                    <SeverityBadge severity="critical" count={data.security.summary.criticalCount} />
                    <SeverityBadge severity="high" count={data.security.summary.highCount} />
                    <SeverityBadge severity="medium" count={data.security.summary.mediumCount} />
                    <SeverityBadge severity="low" count={data.security.summary.lowCount} />
                    {securityTotal === 0 && (
                      <span className="text-muted-foreground text-xs">No issues</span>
                    )}
                  </div>
                </div>
              )}

              {/* Quality Summary */}
              {data?.quality && (
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Code2 className="h-4 w-4 text-blue-500" />
                    <span className="text-sm font-medium">Quality</span>
                    <span className="text-muted-foreground text-xs">({qualityTotal} issues)</span>
                  </div>
                  <div className="flex flex-wrap gap-1">
                    <SeverityBadge severity="critical" count={data.quality.summary.criticalCount} />
                    <SeverityBadge severity="high" count={data.quality.summary.highCount} />
                    <SeverityBadge severity="medium" count={data.quality.summary.mediumCount} />
                    <SeverityBadge severity="low" count={data.quality.summary.lowCount} />
                    {qualityTotal === 0 && (
                      <span className="text-muted-foreground text-xs">No issues</span>
                    )}
                  </div>
                </div>
              )}

              {/* View Details Link */}
              <div className="flex items-center justify-between pt-2 border-t">
                <span className="text-muted-foreground text-xs">
                  {data?.security?.analyzedAt && (
                    <>Analyzed {new Date(data.security.analyzedAt).toLocaleString()}</>
                  )}
                </span>
                <div className="flex items-center gap-1 text-primary text-sm font-medium">
                  View all issues
                  <ArrowRight className="h-4 w-4" />
                </div>
              </div>
            </div>
          ) : (
            <div className="text-muted-foreground py-2 text-center text-sm">
              <p>No analysis reports yet</p>
              <p className="text-xs">Click refresh to run analysis</p>
            </div>
          )}
        </div>
      </div>
    </Link>
  )
}
