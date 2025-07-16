import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { LucideIcon } from 'lucide-react'

interface AuthCardProps {
  title: string
  description?: string
  icon?: LucideIcon
  children: React.ReactNode
}

export function AuthCard({ title, description, icon: Icon, children }: AuthCardProps) {
  return (
    <Card className="w-full">
      <CardHeader className="space-y-4 pb-8 text-center">
        {Icon && (
          <div className="mx-auto w-12 h-12 flex items-center justify-center">
            <Icon className="h-12 w-12 text-primary" strokeWidth={1.5} />
          </div>
        )}
        <div className="space-y-2">
          <CardTitle className="text-3xl font-semibold tracking-tight">
            {title}
          </CardTitle>
          {description && (
            <CardDescription className="text-base text-muted-foreground">{description}</CardDescription>
          )}
        </div>
      </CardHeader>
      <CardContent className="pb-8">{children}</CardContent>
    </Card>
  )
}