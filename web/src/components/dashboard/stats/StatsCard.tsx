import { LucideIcon } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { ReactNode } from 'react';

interface StatsCardProps {
  title: string;
  value: number | string | ReactNode;
  icon: LucideIcon;
  description?: string;
  variant?: 'default' | 'success' | 'warning' | 'error' | 'users' | 'workspaces' | 'environments' | 'services';
  className?: string;
}

const variantStyles = {
  default: 'text-foreground',
  success: 'text-green-500',
  warning: 'text-yellow-500',
  error: 'text-red-500',
  users: 'text-foreground',
  workspaces: 'text-foreground',
  environments: 'text-foreground',
  services: 'text-foreground',
};

const iconVariantStyles = {
  default: 'text-muted-foreground bg-muted',
  success: 'text-green-500 bg-green-500/10',
  warning: 'text-yellow-500 bg-yellow-500/10',
  error: 'text-red-500 bg-red-500/10',
  users: 'text-muted-foreground bg-muted',
  workspaces: 'text-muted-foreground bg-muted',
  environments: 'text-muted-foreground bg-muted',
  services: 'text-muted-foreground bg-muted',
};

export function StatsCard({
  title,
  value,
  icon: Icon,
  description,
  variant = 'default',
  className,
}: StatsCardProps) {
  return (
    <Card className={cn('group transition-colors duration-200 hover:bg-muted/10', className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-base font-semibold text-foreground">
          {title}
        </CardTitle>
        <div className={cn(
          'p-2 rounded-md transition-colors duration-200',
          iconVariantStyles[variant]
        )}>
          <Icon className="h-4 w-4" />
        </div>
      </CardHeader>
      <CardContent>
        <div className={cn(
          'text-2xl font-bold',
          variantStyles[variant]
        )}>
          {value}
        </div>
        {description && (
          <p className="text-xs text-muted-foreground mt-1">
            {description}
          </p>
        )}
      </CardContent>
    </Card>
  );
}