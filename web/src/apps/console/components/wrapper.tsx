import { SubHeader } from '~/components/organisms/sub-header';
import { Link, useSearchParams } from '@remix-run/react';
import { ReactNode } from 'react';
import { IButton } from '~/components/atoms/button';
import { EmptyState } from './empty-state';
import { CustomPagination } from './custom-pagination';

interface WrapperProps {
  children: ReactNode;
  empty?: {
    image?: ReactNode;
    title: string;
    action: IButton;
    is: boolean;
    content: ReactNode;
  };
  header?: {
    title: string;
    backurl?: string;
    action: ReactNode;
  };
  pagination?: any;
}

const Wrapper = ({
  children,
  empty,
  header,
  pagination = null,
}: WrapperProps) => {
  const [sp] = useSearchParams();
  const isEmpty = !(sp.get('search') || sp.get('page')) && empty?.is;
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
      <div className="pt-3xl flex flex-col gap-6xl">
        {!isEmpty && children}
        {!isEmpty && pagination && <CustomPagination pagination={pagination} />}
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
                  <rect y="0.970703" width="226" height="226" fill="#F4F4F5" />
                </svg>
              )
            }
            heading={empty?.title}
            action={empty?.action}
          >
            {empty?.content}
          </EmptyState>
        )}
      </div>
    </>
  );
};

export default Wrapper;
