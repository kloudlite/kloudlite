import { CircleFill, CircleNotch } from '@jengaicons/react';
import { useRevalidator } from '@remix-run/react';
import { cn } from '@kloudlite/design-system/utils';

export const LoadingIndicator = ({
  className,
  size = 1,
}: {
  className?: string;
  size?: number;
}) => {
  return (
    <div
      className={cn(
        'text-text-warning animate-spin flex items-center justify-center aspect-square',
        className
      )}
    >
      <CircleNotch size={16 * size} className="relative" />
      <CircleFill size={8 * size} className="absolute" />
    </div>
  );
};

export const ReloadIndicator = () => {
  const { state } = useRevalidator();

  if (state === 'loading') {
    return (
      <div className="fixed z-20 bottom-lg right-lg">
        <LoadingIndicator className="" size={1} />
      </div>
    );
  }

  return null;
};
