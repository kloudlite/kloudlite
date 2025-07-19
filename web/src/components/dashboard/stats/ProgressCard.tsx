import { LucideIcon } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';

interface ProgressItem {
  label: string;
  value: number;
  max: number;
  color?: 'default' | 'success' | 'warning' | 'error';
}

interface ProgressCardProps {
  title: string;
  icon: LucideIcon;
  items: ProgressItem[];
  className?: string;
}

const progressColors = {
  default: 'bg-primary',
  success: 'bg-green-500',
  warning: 'bg-yellow-500',
  error: 'bg-destructive',
};

function ProgressBar({ item }: { item: ProgressItem }) {
  const percentage = (item.value / item.max) * 100;
  const safePercentage = Math.min(percentage, 100);
  
  const getColor = () => {
    if (item.color) return progressColors[item.color];
    if (percentage > 85) return progressColors.error;
    if (percentage > 70) return progressColors.warning;
    return progressColors.success;
  };

  return (
    <div className="space-y-2">
      <div className="flex justify-between text-sm">
        <span className="text-muted-foreground">{item.label}</span>
        <span className="font-medium">
          {item.value} / {item.max}
        </span>
      </div>
      <div className="w-full bg-muted rounded-full h-2">
        <div
          className={cn('h-2 rounded-full transition-all duration-300', getColor())}
          style={{ width: `${safePercentage}%` }}
        />
      </div>
      <div className="text-xs text-muted-foreground text-right">
        {percentage.toFixed(1)}% utilized
      </div>
    </div>
  );
}

export function ProgressCard({ title, icon: Icon, items, className }: ProgressCardProps) {
  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Icon className="h-5 w-5" />
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {items.map((item, index) => (
          <ProgressBar key={index} item={item} />
        ))}
      </CardContent>
    </Card>
  );
}