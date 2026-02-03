import { EnvironmentSettingsTabs } from '../../_components/environment-settings-tabs'

interface LayoutProps {
  children: React.ReactNode
  params: Promise<{
    id: string
  }>
}

export default async function SettingsLayout({ children, params }: LayoutProps) {
  const { id } = await params
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      {/* Title Section */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">Environment Settings</h1>
        <p className="text-muted-foreground mt-1.5 text-sm">
          Manage environment configuration and access control
        </p>
      </div>

      {/* Tabs Navigation */}
      <EnvironmentSettingsTabs environmentId={id} />

      {/* Content */}
      <div className="mt-6">{children}</div>
    </div>
  )
}
