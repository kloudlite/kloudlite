import { Spinner } from '@jengaicons/react';
import { ReactNode } from 'react';

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
