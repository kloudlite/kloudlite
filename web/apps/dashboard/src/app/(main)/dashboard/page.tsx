import { redirect } from 'next/navigation'
import Link from 'next/link'
import { getSession } from '@/lib/get-session'
import { Button } from '@kloudlite/ui'
import { Cloud, Monitor, Package, Plus, TrendingUp, Activity, Zap, ArrowRight } from 'lucide-react'

export default async function Dashboard() {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  // Mock overview data
  const stats = {
    environments: 3,
    activeEnvironments: 2,
    workspaces: 4,
    runningWorkspaces: 3,
  }

  const statCards = [
    {
      label: 'Total Environments',
      value: stats.environments,
      change: '+2 this week',
      trend: 'up' as const,
      icon: Cloud,
      color: 'from-blue-500 to-cyan-500',
    },
    {
      label: 'Active Now',
      value: stats.activeEnvironments,
      change: `${stats.activeEnvironments}/${stats.environments} active`,
      trend: 'neutral' as const,
      icon: Activity,
      color: 'from-emerald-500 to-teal-500',
    },
    {
      label: 'Total Workspaces',
      value: stats.workspaces,
      change: '+1 this week',
      trend: 'up' as const,
      icon: Monitor,
      color: 'from-violet-500 to-purple-500',
    },
    {
      label: 'Running Now',
      value: stats.runningWorkspaces,
      change: `${stats.runningWorkspaces}/${stats.workspaces} running`,
      trend: 'neutral' as const,
      icon: Zap,
      color: 'from-amber-500 to-orange-500',
    },
  ]

  const resourceCards = [
    {
      title: 'Environments',
      description: 'Create and manage your development environments with powerful isolation and configuration',
      href: '/environments',
      icon: Cloud,
      count: stats.environments,
      gradient: 'from-blue-500/10 to-cyan-500/10',
      borderGradient: 'from-blue-500 to-cyan-500',
      actions: [
        { label: 'View All', href: '/environments' },
        { label: 'Create New', href: '/environments', primary: true },
      ],
    },
    {
      title: 'Workspaces',
      description: 'Manage your application workspaces with integrated development tools and real-time collaboration',
      href: '/workspaces',
      icon: Monitor,
      count: stats.workspaces,
      gradient: 'from-violet-500/10 to-purple-500/10',
      borderGradient: 'from-violet-500 to-purple-500',
      actions: [
        { label: 'View All', href: '/workspaces' },
        { label: 'Create New', href: '/workspaces', primary: true },
      ],
    },
    {
      title: 'Artifacts',
      description: 'Browse and manage container images, packages, and build artifacts for your projects',
      href: '/artifacts',
      icon: Package,
      count: 0,
      gradient: 'from-emerald-500/10 to-teal-500/10',
      borderGradient: 'from-emerald-500 to-teal-500',
      actions: [
        { label: 'Browse', href: '/artifacts' },
      ],
    },
  ]

  return (
    <div className="space-y-8">
      {/* Page Header */}
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">
          Welcome back{session.user?.name ? `, ${session.user.name}` : ''}
        </h1>
        <p className="text-muted-foreground text-sm">
          Here&apos;s what&apos;s happening with your development environment today
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {statCards.map((stat) => {
          const Icon = stat.icon
          return (
            <div
              key={stat.label}
              className="group relative overflow-hidden rounded-xl border bg-card p-6 transition-all duration-300 hover:-translate-y-0.5"
            >
              {/* Gradient background decoration */}
              <div className={`absolute top-0 right-0 h-24 w-24 bg-gradient-to-br ${stat.color} opacity-0 blur-3xl transition-opacity duration-300 group-hover:opacity-20`} />

              <div className="relative space-y-3">
                {/* Icon */}
                <div className={`inline-flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br ${stat.color} opacity-90`}>
                  <Icon className="h-5 w-5 text-white" />
                </div>

                {/* Value */}
                <div>
                  <p className="text-3xl font-bold tracking-tight">{stat.value}</p>
                  <p className="text-muted-foreground mt-1 text-sm font-medium">{stat.label}</p>
                </div>

                {/* Change indicator */}
                <div className="flex items-center gap-1.5">
                  {stat.trend === 'up' && (
                    <TrendingUp className="h-3.5 w-3.5 text-emerald-500" />
                  )}
                  <span className={`text-xs font-medium ${
                    stat.trend === 'up' ? 'text-emerald-600 dark:text-emerald-400' : 'text-muted-foreground'
                  }`}>
                    {stat.change}
                  </span>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Resources Section */}
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold tracking-tight">Your Resources</h2>
            <p className="text-muted-foreground mt-1 text-sm">
              Quick access to your environments, workspaces, and artifacts
            </p>
          </div>
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          {resourceCards.map((resource) => {
            const Icon = resource.icon
            return (
              <div
                key={resource.title}
                className="group relative overflow-hidden rounded-xl border bg-card transition-all duration-300"
              >
                {/* Gradient border effect on hover */}
                <div className={`absolute inset-0 bg-gradient-to-br ${resource.borderGradient} opacity-0 group-hover:opacity-100 transition-opacity duration-300`}
                     style={{ padding: '1px' }}>
                  <div className="h-full w-full rounded-xl bg-card" />
                </div>

                <div className="relative p-6 space-y-4">
                  {/* Header */}
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <div className={`flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br ${resource.gradient} border border-border/50`}>
                        <Icon className="h-6 w-6" />
                      </div>
                      <div>
                        <h3 className="text-lg font-semibold">{resource.title}</h3>
                        <p className="text-muted-foreground text-sm">{resource.count} total</p>
                      </div>
                    </div>
                    <ArrowRight className="h-5 w-5 text-muted-foreground transition-transform duration-300 group-hover:translate-x-1 group-hover:text-foreground" />
                  </div>

                  {/* Description */}
                  <p className="text-muted-foreground text-sm leading-relaxed">
                    {resource.description}
                  </p>

                  {/* Actions */}
                  <div className="flex items-center gap-3 pt-2">
                    {resource.actions.map((action) => (
                      <Link key={action.label} href={action.href}>
                        <Button
                          size="sm"
                          variant={action.primary ? 'default' : 'outline'}
                          className="gap-1.5"
                        >
                          {action.primary && <Plus className="h-3.5 w-3.5" />}
                          {action.label}
                        </Button>
                      </Link>
                    ))}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Quick Actions */}
      <div className="rounded-xl border bg-card p-6">
        <div className="flex items-start justify-between gap-6">
          <div className="flex-1 space-y-2">
            <h3 className="text-lg font-semibold">Get Started</h3>
            <p className="text-muted-foreground text-sm leading-relaxed">
              Create your first environment or workspace to start building and deploying your applications
            </p>
          </div>
          <div className="flex shrink-0 items-center gap-3">
            <Link href="/environments">
              <Button className="gap-2">
                <Plus className="h-4 w-4" />
                New Environment
              </Button>
            </Link>
            <Link href="/workspaces">
              <Button variant="outline" className="gap-2">
                <Plus className="h-4 w-4" />
                New Workspace
              </Button>
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}
