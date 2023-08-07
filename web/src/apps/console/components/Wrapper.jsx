import { SubHeader } from '~/components/organisms/sub-header';
import { Link } from '@remix-run/react';
import Pagination from '~/components/molecule/pagination';
import { EmptyState } from './empty-state';

const Wrapper = ({ children, empty, header, pagination }) => {
  return (
    <>
      {header && (
        <SubHeader
          title={header.title}
          backUrl={header.backurl}
          LinkComponent={Link}
          actions={header.action}
        />
      )}
      <div className="pt-3xl flex flex-col gap-6xl">
        {!empty?.is && children}
        {!empty?.is && pagination && (
          <div className="hidden md:flex">
            <Pagination {...pagination} />
          </div>
        )}
        {empty?.is && (
          <EmptyState
            illustration={
              <svg
                width="226"
                height="227"
                viewBox="0 0 226 227"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <rect y="0.970703" width="226" height="226" fill="#F4F4F5" />
              </svg>
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
