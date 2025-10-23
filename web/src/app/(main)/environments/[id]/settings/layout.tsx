import { SettingsSectionNav } from '../../_components/settings-section-nav'

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
      <div className="flex gap-6">
        <SettingsSectionNav environmentId={id} />
        <div className="flex-1">
          {children}
        </div>
      </div>
    </div>
  )
}
