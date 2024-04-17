import { CircleFill, CircleNotch, Spinner } from '~/console/components/icons';
import { ReactNode } from 'react';
import { cn } from '~/components/utils';

export const LoadingPlaceHolder = ({
  height = 100,
  title,
}: {
  height?: number;
  title?: ReactNode;
}) => {
  return (
    <div
      style={{ minHeight: `${height}px` }}
      className="flex flex-col items-center justify-center gap-xl py-2xl"
    >
      <span className="animate-spin">
        <Spinner color="currentColor" weight={2} size={24} />
      </span>
      <span className="text-text-soft bodyMd">{title || 'Loading'}</span>
    </div>
  );
};

export const LoadingIndicator = ({
  className,
  size = 1,
}: {
  className?: string;
  size?: 1 | 2 | 3;
}) => {
  return (
    <div
      className={cn(
        'text-text-warning animate-spin flex items-center justify-center aspect-square',
        className
      )}
    >
      <CircleNotch size={16 * size} />
      <CircleFill size={8 * size} className="absolute" />
    </div>
  );
};
