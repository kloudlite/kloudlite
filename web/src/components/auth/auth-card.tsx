import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface AuthCardProps {
  title: string
  description?: string
  children: React.ReactNode
}

export function AuthCard({ title, description, children }: AuthCardProps) {
  return (
    <Card className="w-full">
      <CardHeader className="space-y-3 pb-8">
        <CardTitle className="text-3xl font-semibold tracking-tight">{title}</CardTitle>
        {description && (
          <CardDescription className="text-base text-muted-foreground">{description}</CardDescription>
        )}
      </CardHeader>
      <CardContent className="pb-8">{children}</CardContent>
    </Card>
  )
}