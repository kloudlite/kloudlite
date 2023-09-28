import { ChevronLeft, ChevronRight, SmileySad } from '@jengaicons/react';
import { ReactNode } from 'react';
import { IconButton } from '~/components/atoms/button';
import { cn } from '~/components/utils';
import NoResultsFound from './no-results-found';

interface IHeader {
  children?: ReactNode;
  onPrev: () => void;
  onNext: () => void;
  hasPrevious: boolean;
  hasNext: boolean;
}
export const DynamicPaginationHeader = ({
  children,
  onNext,
  onPrev,
  hasNext,
  hasPrevious,
}: IHeader) => {
  return (
    <div className="flex flex-row items-center px-xl py-lg bg-surface-basic-subdued rounded-t">
      <div className="text-text-strong bodyMd flex-1">{children}</div>
      <div className="flex flex-row items-center">
        <IconButton
          icon={<ChevronLeft />}
          size="xs"
          variant="plain"
          onClick={() => onPrev()}
          disabled={!hasPrevious}
        />
        <IconButton
          icon={<ChevronRight />}
          size="xs"
          variant="plain"
          onClick={() => onNext()}
          disabled={!hasNext}
        />
      </div>
    </div>
  );
};

interface IDynamicPagination extends IHeader {
  title: ReactNode;
  hasItems: boolean;
  noItemsMessage: ReactNode;
  className?: string;
}
const DynamicPagination = ({
  children,
  hasNext,
  hasPrevious,
  onNext,
  onPrev,
  title,
  hasItems,
  noItemsMessage,
  className,
}: IDynamicPagination) => {
  return (
    <div className={cn('bg-surface-basic-default flex', className)}>
      {hasItems && (
        <div className="w-full flex flex-col">
          <DynamicPaginationHeader
            {...{ hasNext, hasPrevious, onNext, onPrev }}
          >
            {title}
          </DynamicPaginationHeader>
          <div>{children}</div>
        </div>
      )}
      {!hasItems && (
        <div className="h-full flex flex-row items-center justify-center w-full self-end">
          <NoResultsFound
            title={null}
            subtitle={noItemsMessage}
            compact
            image={<SmileySad size={32} weight={1} />}
            shadow={false}
            border={false}
          />
        </div>
      )}
    </div>
  );
};

export default DynamicPagination;
