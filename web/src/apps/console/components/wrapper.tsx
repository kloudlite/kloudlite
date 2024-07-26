import { SmileySad } from '~/console/components/icons';
import { Link, useSearchParams } from '@remix-run/react';
import { ReactNode } from 'react';
import { IButton } from '~/components/atoms/button';
import { SubHeader } from '~/components/organisms/sub-header';
import { cn } from '~/components/utils';
import { CustomPagination } from './custom-pagination';
import { EmptyState } from './empty-state';
import NoResultsFound, { INoResultsFound } from './no-results-found';
import SecondarySubHeader from './secondary-sub-header';
import TooltipV2 from '~/components/atoms/tooltipV2';

interface WrapperProps {
  children?: ReactNode;
  empty?: {
    image?: ReactNode;
    title: string;
    action?: IButton | ReactNode;
    is: boolean;
    content: ReactNode;
  };
  header?: {
    title: ReactNode;
    backurl?: string;
    action?: ReactNode;
  };
  secondaryHeader?: {
    title: ReactNode;
    action?: ReactNode;
  };
  pagination?: {
    pageInfo: {
      endCursor?: string;
      hasNextPage?: boolean;
      hasPrevPage?: boolean;
      startCursor?: string;
    };
    totalCount: number;
  };
  tools?: ReactNode;
  noResultFound?: INoResultsFound;
}

const Wrapper = ({
  children,
  empty,
  header,
  secondaryHeader,
  pagination,
  tools,
  noResultFound,
}: WrapperProps) => {
  const [sp] = useSearchParams();
  const isSearch = sp.get('search') || sp.get('page');
  const isSearchResultEmpty = isSearch && empty?.is;
  const isEmpty = !isSearch && empty?.is;
  return (
    <>
      {header && (
        <SubHeader
          title={header.title}
          backUrl={header.backurl || ''}
          LinkComponent={Link}
          actions={header.action}
        />
      )}
      {secondaryHeader && (
        <div className="pb-6xl">
          <SecondarySubHeader
            title={secondaryHeader.title}
            action={secondaryHeader.action}
          />
        </div>
      )}
      <div className="flex flex-col">
        {!isEmpty && tools}
        <div className={cn('flex flex-col gap-6xl')}>
          {!isEmpty && !isSearchResultEmpty && children}
          {!isEmpty && pagination && (
            <CustomPagination pagination={pagination} />
          )}
          {isEmpty && (
            <EmptyState
              image={
                empty?.image ? (
                  empty.image
                ) : (
                  <svg
                    width="226"
                    height="227"
                    viewBox="0 0 226 227"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <rect
                      y="0.970703"
                      width="226"
                      height="226"
                      className="fill-surface-basic-subdued"
                    />
                  </svg>
                )
              }
              heading={empty?.title}
              action={empty?.action}
            >
              {empty?.content}
            </EmptyState>
          )}
          {isSearchResultEmpty && (
            <NoResultsFound
              title={noResultFound?.title || 'No results found'}
              subtitle={
                noResultFound?.subtitle ||
                'Try changing the filters or search terms for this view.'
              }
              image={noResultFound?.image || <SmileySad size={40} />}
              action={noResultFound?.action}
            />
          )}
        </div>
      </div>
    </>
  );
};

export default Wrapper;
