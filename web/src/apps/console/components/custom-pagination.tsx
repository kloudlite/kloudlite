import { useSearchParams } from '@remix-run/react';
import Pagination from '~/components/molecule/pagination';
import { useState } from 'react';
import {
  decodeUrl,
  encodeUrl,
  useQueryParameters,
} from '~/root/lib/client/hooks/use-search';
import useDebounce from '~/root/lib/client/hooks/use-debounce';

export const CustomPagination = ({ pagination }: { pagination: any }) => {
  const { startCursor, endCursor, hasPreviousPage, hasNextPage } =
    pagination?.pageInfo || {};

  const { totalCount } = pagination || {};

  const [sp] = useSearchParams();

  const [page, setPage] = useState(() => decodeUrl(sp.get('page')) || '');

  const { setQueryParameters } = useQueryParameters();
  const [isFirstTime, setIsFirstTime] = useState(true);

  useDebounce(
    () => {
      if (isFirstTime) {
        setIsFirstTime(false);
        return;
      }
      if (page) {
        setQueryParameters({
          page: encodeUrl(page),
        });
      }
    },
    300,
    [page]
  );

  const newPagination = (k: any) => k;
  return (
    <div className="hidden md:flex">
      <Pagination
        {...pagination}
        showNumbers={false}
        isPrevDisabled={!hasPreviousPage}
        isNextDisabled={!hasNextPage}
        totalItems={totalCount}
        itemsPerPage={10}
        onClickNext={() => {
          if (endCursor) {
            setPage(newPagination({ first: 10, after: endCursor }));
          } else {
            setPage(newPagination({ first: 10 }));
          }
        }}
        onClickPrev={() => {
          if (startCursor) {
            setPage(newPagination({ last: 10, before: startCursor }));
          } else {
            setPage(newPagination({ last: 10 }));
          }
        }}
      />
    </div>
  );
};
