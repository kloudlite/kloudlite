import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { Github, Rss } from 'lucide-react'

// Cross marker component
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
    </div>
  )
}

function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative mx-auto max-w-5xl', className)}>
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />
        <CrossMarker className="top-0 left-0 -translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="top-0 right-0 translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 left-0 -translate-x-1/2 translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 right-0 translate-x-1/2 translate-y-1/2 w-5 h-5" />
      </div>
      <div className="relative">{children}</div>
    </div>
  )
}

interface ChangelogEntry {
  version: string
  date: string
  title: string
  highlights: string[]
  changes: {
    added?: string[]
    changed?: string[]
    fixed?: string[]
  }
}

const changelog: ChangelogEntry[] = [
  {
    version: '0.5.0',
    date: 'Dec 15, 2024',
    title: 'AI-Powered Development',
    highlights: [
      'AI agent integrations for Claude Code, Cursor, and OpenCode',
      'Workspace cloning for parallel development workflows',
    ],
    changes: {
      added: [
        'AI agent integration support with automatic context sharing',
        'Workspace cloning with isolated git worktrees',
        'Automated code scanning and security analysis',
        'New workspace templates for AI-assisted development',
      ],
      changed: [
        'Improved workspace startup time by 40%',
        'Redesigned workspace dashboard with real-time metrics',
      ],
      fixed: [
        'Environment connection stability on network changes',
        'Package cache invalidation issues',
      ],
    },
  },
  {
    version: '0.4.0',
    date: 'Nov 20, 2024',
    title: 'Service Intercepts',
    highlights: [
      'Route production traffic to your workspace for real-time testing',
      'Multi-port forwarding with automatic DNS resolution',
    ],
    changes: {
      added: [
        'Service intercept functionality with one-click setup',
        'Multi-port forwarding support',
        'Intercept status dashboard with live traffic metrics',
        'Automatic SSL certificate handling for intercepted services',
      ],
      changed: [
        'Redesigned workspace connection flow',
        'Improved DNS resolution in connected environments',
      ],
      fixed: [
        'WebSocket connections in intercepted services',
        'Timeout handling for long-running requests',
      ],
    },
  },
  {
    version: '0.3.0',
    date: 'Oct 10, 2024',
    title: 'Environment Cloning',
    highlights: [
      'Clone entire environments with all services and configs',
      'Share environments with team members instantly',
    ],
    changes: {
      added: [
        'One-click environment cloning',
        'Selective resource cloning options',
        'Environment sharing with granular permissions',
        'Environment comparison view',
      ],
      changed: [
        'Improved secrets management UI',
        'Better config sync performance',
      ],
      fixed: [
        'Config sync delays in large environments',
        'Secret rotation edge cases',
      ],
    },
  },
  {
    version: '0.2.0',
    date: 'Sep 5, 2024',
    title: 'Package Management',
    highlights: [
      'Nix-based package management for reproducible environments',
      'CLI-driven package installation and updates',
    ],
    changes: {
      added: [
        'Nix package manager integration',
        'Package search and installation via CLI',
        'Environment-specific package profiles',
        'Package version pinning',
      ],
      changed: [
        'Faster workspace provisioning',
        'Improved package resolution algorithm',
      ],
      fixed: [
        'Package version conflicts',
        'Cache corruption on interrupted installs',
      ],
    },
  },
  {
    version: '0.1.0',
    date: 'Aug 1, 2024',
    title: 'Initial Release',
    highlights: [
      'Cloud development workspaces with VS Code integration',
      'Environment management and service orchestration',
    ],
    changes: {
      added: [
        'Cloud development workspaces',
        'Environment management',
        'VS Code web integration',
        'SSH access to workspaces',
        'CLI tool for workspace operations',
      ],
    },
  },
]

export default function ChangelogPage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="changelog" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-24">
                <div className="text-center">
                  <h1 className="text-[2.5rem] font-bold leading-[1.08] tracking-[-0.035em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                    <span className="text-foreground/40">C</span><span className="text-foreground">hangelog</span>
                  </h1>
                  <p className="text-foreground/55 mx-auto mt-6 max-w-md text-lg leading-relaxed">
                    New updates and improvements to Kloudlite.
                  </p>
                  <div className="mt-8 flex items-center justify-center gap-3">
                    <a
                      href="https://github.com/kloudlite/kloudlite/releases"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-2 px-4 py-2 border border-foreground/10 text-foreground/60 hover:text-foreground hover:border-foreground/20 text-sm transition-colors"
                    >
                      <Github className="h-4 w-4" />
                      Releases
                    </a>
                    <a
                      href="/changelog/rss"
                      className="inline-flex items-center gap-2 px-4 py-2 border border-foreground/10 text-foreground/60 hover:text-foreground hover:border-foreground/20 text-sm transition-colors"
                    >
                      <Rss className="h-4 w-4" />
                      RSS
                    </a>
                  </div>
                </div>
              </div>

              {/* Nightly - What's Cooking */}
              <div className="-mx-6 lg:-mx-12 border-t border-b border-foreground/10 bg-gradient-to-b from-primary/[0.03] to-transparent">
                <div className="p-8 lg:p-10 border-b border-foreground/10">
                  <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
                    <div>
                      <div className="flex items-center gap-4">
                        <span className="text-foreground text-2xl lg:text-3xl font-bold tracking-tight font-mono">
                          Nightly
                        </span>
                        <span className="px-2 py-0.5 bg-amber-500/10 text-amber-600 dark:text-amber-400 text-xs font-medium animate-pulse">
                          In Progress
                        </span>
                      </div>
                      <h2 className="text-foreground mt-2 text-xl font-semibold">What&apos;s Cooking</h2>
                    </div>
                    <a
                      href="https://github.com/kloudlite/kloudlite/tree/development"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-foreground/40 hover:text-foreground text-sm transition-colors"
                    >
                      View development branch →
                    </a>
                  </div>

                  <p className="text-foreground/50 mt-4 text-sm leading-relaxed max-w-2xl">
                    Features currently in development. These may change before release.
                    Try them out on the nightly build and share your feedback.
                  </p>
                </div>

                <div className="grid lg:grid-cols-2 divide-y lg:divide-y-0 lg:divide-x divide-foreground/10">
                  <div className="p-8 lg:p-10">
                    <h3 className="text-foreground/40 text-xs font-semibold uppercase tracking-wider mb-4">
                      Coming Soon
                    </h3>
                    <ul className="space-y-4">
                      <li className="flex items-start gap-3">
                        <span className="w-1.5 h-1.5 rounded-full bg-primary mt-2 flex-shrink-0" />
                        <div>
                          <p className="text-foreground font-medium text-sm">GPU-enabled Workspaces</p>
                          <p className="text-foreground/50 text-sm mt-1">Run ML workloads and AI models directly in your workspace</p>
                        </div>
                      </li>
                      <li className="flex items-start gap-3">
                        <span className="w-1.5 h-1.5 rounded-full bg-primary mt-2 flex-shrink-0" />
                        <div>
                          <p className="text-foreground font-medium text-sm">Team Collaboration</p>
                          <p className="text-foreground/50 text-sm mt-1">Real-time pair programming and workspace sharing</p>
                        </div>
                      </li>
                      <li className="flex items-start gap-3">
                        <span className="w-1.5 h-1.5 rounded-full bg-primary mt-2 flex-shrink-0" />
                        <div>
                          <p className="text-foreground font-medium text-sm">Prebuilt Images</p>
                          <p className="text-foreground/50 text-sm mt-1">One-click workspace templates for popular frameworks</p>
                        </div>
                      </li>
                    </ul>
                  </div>

                  <div className="p-8 lg:p-10">
                    <h3 className="text-foreground/40 text-xs font-semibold uppercase tracking-wider mb-4">
                      In Development
                    </h3>
                    <ul className="space-y-4">
                      <li className="flex items-start gap-3">
                        <span className="w-1.5 h-1.5 rounded-full bg-amber-500 mt-2 flex-shrink-0 animate-pulse" />
                        <div>
                          <p className="text-foreground font-medium text-sm">Multi-cluster Support</p>
                          <p className="text-foreground/50 text-sm mt-1">Connect workspaces across multiple Kubernetes clusters</p>
                        </div>
                      </li>
                      <li className="flex items-start gap-3">
                        <span className="w-1.5 h-1.5 rounded-full bg-amber-500 mt-2 flex-shrink-0 animate-pulse" />
                        <div>
                          <p className="text-foreground font-medium text-sm">Database Snapshots</p>
                          <p className="text-foreground/50 text-sm mt-1">Instant database cloning for isolated testing</p>
                        </div>
                      </li>
                      <li className="flex items-start gap-3">
                        <span className="w-1.5 h-1.5 rounded-full bg-amber-500 mt-2 flex-shrink-0 animate-pulse" />
                        <div>
                          <p className="text-foreground font-medium text-sm">VS Code Desktop Extension</p>
                          <p className="text-foreground/50 text-sm mt-1">Native VS Code integration with remote workspaces</p>
                        </div>
                      </li>
                    </ul>
                  </div>
                </div>
              </div>

              {/* Section Divider */}
              <div className="-mx-6 lg:-mx-12 p-6">
                <p className="text-foreground/30 text-xs font-semibold uppercase tracking-wider text-center">
                  Released
                </p>
              </div>

              {/* Released Versions */}
              <div className="-mx-6 lg:-mx-12 border-t border-foreground/10">
                {changelog.map((entry, index) => (
                  <div
                    key={entry.version}
                    className={cn(
                      'border-b border-foreground/10',
                      index === 0 && 'bg-foreground/[0.01]'
                    )}
                  >
                    {/* Entry Header */}
                    <div className="p-8 lg:p-10 border-b border-foreground/10">
                      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
                        <div>
                          <div className="flex items-center gap-4">
                            <span className="text-foreground text-2xl lg:text-3xl font-bold tracking-tight font-mono">
                              {entry.version}
                            </span>
                            {index === 0 && (
                              <span className="px-2 py-0.5 bg-primary/10 text-primary text-xs font-medium">
                                Latest
                              </span>
                            )}
                          </div>
                          <h2 className="text-foreground mt-2 text-xl font-semibold">{entry.title}</h2>
                        </div>
                        <time className="text-foreground/40 text-sm font-mono">{entry.date}</time>
                      </div>

                      {/* Highlights */}
                      <div className="mt-6 flex flex-wrap gap-3">
                        {entry.highlights.map((highlight, i) => (
                          <span
                            key={i}
                            className="px-3 py-1.5 bg-foreground/5 text-foreground/70 text-sm"
                          >
                            {highlight}
                          </span>
                        ))}
                      </div>
                    </div>

                    {/* Changes Grid */}
                    <div className="grid lg:grid-cols-3 divide-y lg:divide-y-0 lg:divide-x divide-foreground/10">
                      {entry.changes.added && entry.changes.added.length > 0 && (
                        <div className="p-8 lg:p-10">
                          <h3 className="text-green-600 dark:text-green-400 text-xs font-semibold uppercase tracking-wider mb-4">
                            Added
                          </h3>
                          <ul className="space-y-3">
                            {entry.changes.added.map((item, i) => (
                              <li key={i} className="text-foreground/60 text-sm leading-relaxed flex items-start gap-2">
                                <span className="text-green-600 dark:text-green-400 mt-1.5">+</span>
                                {item}
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}

                      {entry.changes.changed && entry.changes.changed.length > 0 && (
                        <div className="p-8 lg:p-10">
                          <h3 className="text-blue-600 dark:text-blue-400 text-xs font-semibold uppercase tracking-wider mb-4">
                            Changed
                          </h3>
                          <ul className="space-y-3">
                            {entry.changes.changed.map((item, i) => (
                              <li key={i} className="text-foreground/60 text-sm leading-relaxed flex items-start gap-2">
                                <span className="text-blue-600 dark:text-blue-400 mt-1.5">~</span>
                                {item}
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}

                      {entry.changes.fixed && entry.changes.fixed.length > 0 && (
                        <div className="p-8 lg:p-10">
                          <h3 className="text-amber-600 dark:text-amber-400 text-xs font-semibold uppercase tracking-wider mb-4">
                            Fixed
                          </h3>
                          <ul className="space-y-3">
                            {entry.changes.fixed.map((item, i) => (
                              <li key={i} className="text-foreground/60 text-sm leading-relaxed flex items-start gap-2">
                                <span className="text-amber-600 dark:text-amber-400 mt-1.5">*</span>
                                {item}
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}

                      {/* Fill empty columns for initial release */}
                      {!entry.changes.changed && !entry.changes.fixed && (
                        <>
                          <div className="hidden lg:block p-8 lg:p-10" />
                          <div className="hidden lg:block p-8 lg:p-10" />
                        </>
                      )}
                    </div>
                  </div>
                ))}
              </div>

              {/* Footer */}
              <div className="p-8 lg:p-10 -mx-6 lg:-mx-12 flex items-center justify-center">
                <p className="text-foreground/40 text-sm text-center">
                  Looking for older releases?{' '}
                  <a
                    href="https://github.com/kloudlite/kloudlite/releases"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline"
                  >
                    View all releases on GitHub
                  </a>
                </p>
              </div>
            </GridContainer>
          </div>

          <WebsiteFooter />
        </main>
      </ScrollArea>
    </div>
  )
}
