interface LayoutProps {
  children: React.ReactNode
  params: Promise<{
    id: string
  }>
}

export default async function ServicesLayout({ children }: LayoutProps) {
  return <div className="pt-6">{children}</div>
}
