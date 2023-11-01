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
  className?: string;
}
export const DynamicPaginationHeader = ({
  children,
  onNext,
  onPrev,
  hasNext,
  hasPrevious,
  className,
}: IHeader) => {
  return (
    <div
      className={cn(
        'flex flex-row items-center pr-xl border-b border-border-disabled',
        {
          'flex flex-row items-center px-xl py-lg bg-surface-basic-subdued rounded-t':
            typeof children === 'string',
        },
        className
      )}
    >
      <div
        className={cn('flex-1', {
          'text-text-strong bodyMd flex-1': typeof children === 'string',
        })}
      >
        {children}
      </div>
      <div className="flex flex-row items-center">
        <IconButton
          icon={<ChevronLeft />}
          size="xs"
          variant="plain"
          onClick={() => onPrev()}
          disabled={!hasPrevious}
          className={cn({
            invisible: !hasPrevious,
          })}
        />
        <IconButton
          icon={<ChevronRight />}
          size="xs"
          variant="plain"
          onClick={() => onNext()}
          disabled={!hasNext}
          className={cn({
            invisible: !hasNext,
          })}
        />
      </div>
    </div>
  );
};

interface IDynamicPagination extends IHeader {
  header: ReactNode;
  hasItems: boolean;
  noItemsMessage: ReactNode;
  className?: string;
  headerClassName?: string;
}
const DynamicPagination = ({
  children,
  hasNext,
  hasPrevious,
  onNext,
  onPrev,
  header,
  hasItems,
  noItemsMessage,
  className,
  headerClassName,
}: IDynamicPagination) => {
  return (
    <div className={cn('bg-surface-basic-default flex', className)}>
      {hasItems && (
        <div className="w-full flex flex-col">
          <DynamicPaginationHeader
            {...{
              hasNext,
              hasPrevious,
              onNext,
              onPrev,
              className: headerClassName,
            }}
          >
            {header}
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
