'use client'

import { useEffect, useState, useCallback, useMemo } from 'react'
import { useParams } from 'next/navigation'
import {
  Shield,
  Code2,
  RefreshCw,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Loader2,
  ChevronDown,
  ChevronRight,
  Search,
  FileCode,
  ArrowLeft,
} from 'lucide-react'
import Link from 'next/link'
import { Button } from '@kloudlite/ui'
import { Breadcrumb } from '@/components/breadcrumb'
import {
  getCodeAnalysis,
  triggerCodeAnalysis,
  type CodeAnalysisResponse,
  type CodeAnalysisFinding,
} from '@/app/actions/workspace.actions'

type SeverityFilter = 'all' | 'critical' | 'high' | 'medium' | 'low'
type TypeFilter = 'all' | 'security' | 'quality'

const severityColors: Record<string, string> = {
  critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  high: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
  medium: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
  low: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
}

const severityOrder: Record<string, number> = {
  critical: 0,
  high: 1,
  medium: 2,
  low: 3,
}

function SeverityBadge({ severity }: { severity: string }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${severityColors[severity] || 'bg-secondary text-secondary-foreground'}`}
    >
      {severity}
    </span>
  )
}

function ScoreIndicator({ score, label }: { score: number; label: string }) {
  let color = 'text-green-600 dark:text-green-400'
  let bgColor = 'bg-green-100 dark:bg-green-900/30'
  let borderColor = 'border-green-200 dark:border-green-800'
  let Icon = CheckCircle2

  if (score < 50) {
    color = 'text-red-600 dark:text-red-400'
    bgColor = 'bg-red-100 dark:bg-red-900/30'
    borderColor = 'border-red-200 dark:border-red-800'
    Icon = XCircle
  } else if (score < 80) {
    color = 'text-yellow-600 dark:text-yellow-400'
    bgColor = 'bg-yellow-100 dark:bg-yellow-900/30'
    borderColor = 'border-yellow-200 dark:border-yellow-800'
    Icon = AlertTriangle
  }

  return (
    <div className={`rounded-lg border ${borderColor} ${bgColor} p-4`}>
      <div className="flex items-center justify-between">
        <span className="text-muted-foreground text-sm font-medium">{label}</span>
        <Icon className={`h-5 w-5 ${color}`} />
      </div>
      <div className="mt-2 flex items-baseline gap-1">
        <span className={`text-3xl font-bold ${color}`}>{score}</span>
        <span className="text-muted-foreground text-sm">/100</span>
      </div>
    </div>
  )
}

function SummaryCard({
  title,
  icon: Icon,
  counts,
  type,
}: {
  title: string
  icon: typeof Shield
  counts: { critical: number; high: number; medium: number; low: number }
  type: 'security' | 'quality'
}) {
  const total = counts.critical + counts.high + counts.medium + counts.low
  const iconBgColor = type === 'security' ? 'bg-red-100 dark:bg-red-900/30' : 'bg-blue-100 dark:bg-blue-900/30'
  const iconColor = type === 'security' ? 'text-red-600 dark:text-red-400' : 'text-blue-600 dark:text-blue-400'

  return (
    <div className="bg-card rounded-lg border p-4">
      <div className="flex items-center gap-3">
        <div className={`rounded-lg p-2 ${iconBgColor}`}>
          <Icon className={`h-5 w-5 ${iconColor}`} />
        </div>
        <div>
          <h3 className="font-medium">{title}</h3>
          <p className="text-muted-foreground text-sm">{total} issues found</p>
        </div>
      </div>
      <div className="mt-4 grid grid-cols-4 gap-2 text-center">
        <div>
          <div className="text-lg font-semibold text-red-600 dark:text-red-400">{counts.critical}</div>
          <div className="text-muted-foreground text-xs">Critical</div>
        </div>
        <div>
          <div className="text-lg font-semibold text-orange-600 dark:text-orange-400">{counts.high}</div>
          <div className="text-muted-foreground text-xs">High</div>
        </div>
        <div>
          <div className="text-lg font-semibold text-yellow-600 dark:text-yellow-400">{counts.medium}</div>
          <div className="text-muted-foreground text-xs">Medium</div>
        </div>
        <div>
          <div className="text-lg font-semibold text-blue-600 dark:text-blue-400">{counts.low}</div>
          <div className="text-muted-foreground text-xs">Low</div>
        </div>
      </div>
    </div>
  )
}

interface FindingWithType extends CodeAnalysisFinding {
  type: 'security' | 'quality'
  id: string
}

function getShortTitle(title: string): string {
  // Extract a short title from the description - take first sentence or first 80 chars
  const firstSentence = title.split(/[.!?]\s/)[0]
  if (firstSentence.length <= 80) {
    return firstSentence + (title.length > firstSentence.length ? '...' : '')
  }
  return title.substring(0, 77) + '...'
}

function getRelativePath(filePath: string): string {
  // Extract path relative to workspace folder
  // Paths look like: /var/lib/kloudlite/home/workspaces/{workspace}/path/to/file
  const workspacesMatch = filePath.match(/\/workspaces\/[^/]+\/(.+)$/)
  if (workspacesMatch) {
    return workspacesMatch[1]
  }
  // Fallback: just return the filename
  const parts = filePath.split('/')
  return parts[parts.length - 1]
}

function FindingRow({ finding, isExpanded, onToggle }: { finding: FindingWithType; isExpanded: boolean; onToggle: () => void }) {
  const shortTitle = getShortTitle(finding.title)
  const relativePath = getRelativePath(finding.file)

  return (
    <>
      <tr
        className="hover:bg-muted/50 cursor-pointer transition-colors"
        onClick={onToggle}
      >
        <td className="w-10 px-4 py-3">
          <button className="text-muted-foreground hover:text-foreground">
            {isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
          </button>
        </td>
        <td className="px-4 py-3">
          <div className="flex items-start gap-2">
            <div className="mt-0.5 flex-shrink-0">
              {finding.type === 'security' ? (
                <Shield className="h-4 w-4 text-red-500" />
              ) : (
                <Code2 className="h-4 w-4 text-blue-500" />
              )}
            </div>
            <span className="font-medium text-sm leading-snug">{shortTitle}</span>
          </div>
        </td>
        <td className="w-28 px-4 py-3">
          <span className="text-muted-foreground text-sm">{finding.category}</span>
        </td>
        <td className="w-56 px-4 py-3">
          <div className="flex items-center gap-1" title={relativePath}>
            <FileCode className="text-muted-foreground h-3.5 w-3.5 flex-shrink-0" />
            <span className="text-muted-foreground font-mono text-xs truncate max-w-[200px]">
              {relativePath}
              {finding.line ? `:${finding.line}` : ''}
            </span>
          </div>
        </td>
        <td className="w-24 px-4 py-3">
          <SeverityBadge severity={finding.severity} />
        </td>
      </tr>
      {isExpanded && (
        <tr className="bg-muted/30">
          <td colSpan={5} className="px-4 py-4">
            <div className="ml-8 space-y-4">
              <div>
                <h4 className="text-sm font-medium mb-1">Description</h4>
                <p className="text-muted-foreground text-sm whitespace-pre-wrap">{finding.description || finding.title}</p>
              </div>
              {finding.recommendation && (
                <div>
                  <h4 className="text-sm font-medium mb-1">Recommendation</h4>
                  <p className="text-muted-foreground text-sm whitespace-pre-wrap">{finding.recommendation}</p>
                </div>
              )}
              <div>
                <h4 className="text-sm font-medium mb-1">Location</h4>
                <p className="text-muted-foreground font-mono text-xs bg-muted/50 px-2 py-1 rounded inline-block">
                  {relativePath}{finding.line ? `:${finding.line}` : ''}
                </p>
              </div>
            </div>
          </td>
        </tr>
      )}
    </>
  )
}

export default function CodeAnalysisPage() {
  const params = useParams()
  const namespace = params.namespace as string
  const workspaceName = params.name as string

  const [data, setData] = useState<CodeAnalysisResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [triggering, setTriggering] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Filter states
  const [typeFilter, setTypeFilter] = useState<TypeFilter>('all')
  const [severityFilter, setSeverityFilter] = useState<SeverityFilter>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set())

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

  const handleTriggerAnalysis = async () => {
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

  // Combine findings from both reports
  const allFindings = useMemo((): FindingWithType[] => {
    const findings: FindingWithType[] = []

    if (data?.security?.findings) {
      data.security.findings.forEach((f, idx) => {
        findings.push({ ...f, type: 'security', id: `security-${idx}` })
      })
    }

    if (data?.quality?.findings) {
      data.quality.findings.forEach((f, idx) => {
        findings.push({ ...f, type: 'quality', id: `quality-${idx}` })
      })
    }

    // Sort by severity
    findings.sort((a, b) => (severityOrder[a.severity] ?? 4) - (severityOrder[b.severity] ?? 4))

    return findings
  }, [data])

  // Filter findings
  const filteredFindings = useMemo(() => {
    return allFindings.filter((finding) => {
      // Type filter
      if (typeFilter !== 'all' && finding.type !== typeFilter) {
        return false
      }

      // Severity filter
      if (severityFilter !== 'all' && finding.severity !== severityFilter) {
        return false
      }

      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase()
        return (
          finding.title.toLowerCase().includes(query) ||
          finding.file.toLowerCase().includes(query) ||
          finding.category.toLowerCase().includes(query) ||
          finding.description.toLowerCase().includes(query)
        )
      }

      return true
    })
  }, [allFindings, typeFilter, severityFilter, searchQuery])

  const toggleRow = (id: string) => {
    setExpandedRows((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const breadcrumbItems = [
    { label: 'Workspaces', href: '/workspaces' },
    { label: workspaceName, href: `/workspaces/${namespace}/${workspaceName}` },
    { label: 'Code Analysis' },
  ]

  const isAnalyzing = data?.status?.inProgress || data?.status?.pendingAnalysis
  const hasReports = data?.security || data?.quality

  // Calculate summary counts
  const securityCounts = {
    critical: data?.security?.summary?.criticalCount || 0,
    high: data?.security?.summary?.highCount || 0,
    medium: data?.security?.summary?.mediumCount || 0,
    low: data?.security?.summary?.lowCount || 0,
  }

  const qualityCounts = {
    critical: data?.quality?.summary?.criticalCount || 0,
    high: data?.quality?.summary?.highCount || 0,
    medium: data?.quality?.summary?.mediumCount || 0,
    low: data?.quality?.summary?.lowCount || 0,
  }

  if (loading) {
    return (
      <div className="flex min-h-[400px] items-center justify-center">
        <div className="flex items-center gap-2">
          <Loader2 className="text-muted-foreground h-5 w-5 animate-spin" />
          <span className="text-muted-foreground">Loading analysis...</span>
        </div>
      </div>
    )
  }

  return (
    <>
      {/* Header */}
      <div className="bg-card border-b">
        <div className="mx-auto max-w-7xl px-6">
          <div className="py-4">
            <Breadcrumb items={breadcrumbItems} />
          </div>

          <div className="flex items-center justify-between pb-6">
            <div>
              <div className="flex items-center gap-3">
                <Link
                  href={`/workspaces/${namespace}/${workspaceName}`}
                  className="text-muted-foreground hover:text-foreground rounded-md p-1 transition-colors"
                >
                  <ArrowLeft className="h-5 w-5" />
                </Link>
                <h1 className="text-2xl font-semibold">Code Analysis</h1>
              </div>
              {hasReports && data?.security?.analyzedAt && (
                <p className="text-muted-foreground ml-9 mt-1 text-sm">
                  Last analyzed: {new Date(data.security.analyzedAt).toLocaleString()}
                </p>
              )}
            </div>
            <Button onClick={handleTriggerAnalysis} disabled={triggering || isAnalyzing}>
              {triggering || isAnalyzing ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <RefreshCw className="mr-2 h-4 w-4" />
              )}
              {isAnalyzing ? 'Analyzing...' : 'Re-analyze'}
            </Button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="mx-auto max-w-7xl px-6 py-8">
        {error && !hasReports ? (
          <div className="bg-card rounded-lg border p-12 text-center">
            <AlertTriangle className="text-muted-foreground mx-auto h-12 w-12" />
            <h3 className="mt-4 text-lg font-medium">No Analysis Available</h3>
            <p className="text-muted-foreground mt-2">
              Code analysis has not been run yet or is not available for this workspace.
            </p>
            <Button onClick={handleTriggerAnalysis} className="mt-4" disabled={triggering}>
              {triggering ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCw className="mr-2 h-4 w-4" />}
              Run Analysis
            </Button>
          </div>
        ) : (
          <>
            {/* Summary Cards */}
            <div className="mb-8 grid gap-4 md:grid-cols-3">
              <SummaryCard title="Security Issues" icon={Shield} counts={securityCounts} type="security" />
              <SummaryCard title="Code Quality" icon={Code2} counts={qualityCounts} type="quality" />
              <ScoreIndicator score={data?.quality?.summary?.score || 0} label="Overall Score" />
            </div>

            {/* Filters */}
            <div className="mb-6 flex flex-wrap items-center gap-4">
              {/* Type Filter */}
              <div className="bg-muted flex items-center gap-1 rounded-md p-1">
                {(['all', 'security', 'quality'] as TypeFilter[]).map((type) => (
                  <button
                    key={type}
                    onClick={() => setTypeFilter(type)}
                    className={`rounded px-3 py-1 text-sm transition-colors ${
                      typeFilter === type
                        ? 'bg-background text-foreground
                        : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    {type === 'all' ? 'All Types' : type.charAt(0).toUpperCase() + type.slice(1)}
                  </button>
                ))}
              </div>

              {/* Severity Filter */}
              <div className="bg-muted flex items-center gap-1 rounded-md p-1">
                {(['all', 'critical', 'high', 'medium', 'low'] as SeverityFilter[]).map((severity) => (
                  <button
                    key={severity}
                    onClick={() => setSeverityFilter(severity)}
                    className={`rounded px-3 py-1 text-sm transition-colors ${
                      severityFilter === severity
                        ? 'bg-background text-foreground
                        : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    {severity === 'all' ? 'All Severities' : severity.charAt(0).toUpperCase() + severity.slice(1)}
                  </button>
                ))}
              </div>

              {/* Search */}
              <div className="relative flex-1 min-w-[200px]">
                <Search className="text-muted-foreground absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" />
                <input
                  type="text"
                  placeholder="Search by file, title, or category..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="bg-background border-input placeholder:text-muted-foreground focus-visible:ring-ring w-full rounded-md border py-2 pl-9 pr-4 text-sm focus-visible:outline-none focus-visible:ring-1"
                />
              </div>
            </div>

            {/* Findings Table */}
            <div className="bg-card overflow-hidden rounded-lg border">
              <div className="overflow-x-auto">
                <table className="min-w-full table-fixed">
                  <thead className="bg-muted/50 border-b">
                    <tr>
                      <th className="w-10 px-4 py-3"></th>
                      <th className="text-muted-foreground px-4 py-3 text-left text-xs font-medium uppercase tracking-wider">
                        Issue
                      </th>
                      <th className="text-muted-foreground w-28 px-4 py-3 text-left text-xs font-medium uppercase tracking-wider">
                        Category
                      </th>
                      <th className="text-muted-foreground w-56 px-4 py-3 text-left text-xs font-medium uppercase tracking-wider">
                        Location
                      </th>
                      <th className="text-muted-foreground w-24 px-4 py-3 text-left text-xs font-medium uppercase tracking-wider">
                        Severity
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y">
                    {filteredFindings.length === 0 ? (
                      <tr>
                        <td colSpan={5} className="px-4 py-12 text-center">
                          <div className="text-muted-foreground">
                            {allFindings.length === 0 ? (
                              <>
                                <CheckCircle2 className="mx-auto h-12 w-12 text-green-500" />
                                <p className="mt-4 text-lg font-medium text-green-600 dark:text-green-400">
                                  No issues found
                                </p>
                                <p className="mt-1">Your code looks good!</p>
                              </>
                            ) : (
                              <>
                                <Search className="mx-auto h-12 w-12" />
                                <p className="mt-4">No findings match your filters</p>
                              </>
                            )}
                          </div>
                        </td>
                      </tr>
                    ) : (
                      filteredFindings.map((finding) => (
                        <FindingRow
                          key={finding.id}
                          finding={finding}
                          isExpanded={expandedRows.has(finding.id)}
                          onToggle={() => toggleRow(finding.id)}
                        />
                      ))
                    )}
                  </tbody>
                </table>
              </div>
              {filteredFindings.length > 0 && (
                <div className="text-muted-foreground border-t px-4 py-3 text-sm">
                  Showing {filteredFindings.length} of {allFindings.length} issues
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </>
  )
}
