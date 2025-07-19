import { cn } from '@/lib/utils';

interface StatusDotProps {
  status: 'online' | 'offline' | 'healthy' | 'degraded' | 'failed';
  className?: string;
}

const statusVariants = {
  online: 'bg-green-500',
  offline: 'bg-muted',
  healthy: 'bg-green-500', 
  degraded: 'bg-yellow-500',
  failed: 'bg-destructive',
};

export function StatusDot({ status, className }: StatusDotProps) {
  return (
    <div className={cn(
      'w-2 h-2 rounded-full',
      statusVariants[status],
      className
    )} />
  );
}