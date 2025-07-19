import { cn } from '@/lib/utils';

interface CountDisplayProps {
  active: number;
  total: number;
  className?: string;
}

export function CountDisplay({ active, total, className }: CountDisplayProps) {
  return (
    <span className={cn('text-sm font-medium', className)}>
      <span className="text-green-500">{active}</span>
      <span className="text-muted-foreground">/{total}</span>
    </span>
  );
}