import { LucideIcon } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';

interface StatsSectionProps {
  title: string;
  description?: string;
  icon?: LucideIcon;
  className?: string;
  children: React.ReactNode;
  variant?: 'default' | 'compact';
}

export function StatsSection({
  title,
  description,
  icon: Icon,
  className,
  children,
  variant = 'default',
}: StatsSectionProps) {
  if (variant === 'compact') {
    return (
      <div className={cn('space-y-4', className)}>
        <div className="flex items-center gap-2">
          {Icon && <Icon className="h-5 w-5 text-muted-foreground" />}
          <h3 className="text-lg font-semibold">{title}</h3>
        </div>
        <div className="grid gap-4">{children}</div>
      </div>
    );
  }

  return (
    <Card className={cn('', className)}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {Icon && <Icon className="h-5 w-5" />}
          {title}
        </CardTitle>
        {description && (
          <p className="text-sm text-muted-foreground">{description}</p>
        )}
      </CardHeader>
      <CardContent>
        <div className="grid gap-4">{children}</div>
      </CardContent>
    </Card>
  );
}