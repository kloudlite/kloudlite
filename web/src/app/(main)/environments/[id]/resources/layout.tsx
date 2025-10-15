import { ResourceSectionNav } from '../../_components/resource-section-nav'

interface LayoutProps {
  children: React.ReactNode
  params: {
    id: string
  }
}

export default async function ResourcesLayout({ children, params }: LayoutProps) {
  const { id } = await params
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      <div className="flex gap-6">
        <ResourceSectionNav environmentId={id} />
        <div className="flex-1">
          {children}
        </div>
      </div>
    </div>
  )
}
