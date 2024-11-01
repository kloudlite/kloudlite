import { useSearchParams } from '@remix-run/react';
import { useState } from 'react';
import Pagination from '@kloudlite/design-system/molecule/pagination';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import {
  decodeUrl,
  encodeUrl,
  useQueryParameters,
} from '~/root/lib/client/hooks/use-search';

export const CustomPagination = ({ pagination }: { pagination: any }) => {
  const { startCursor, endCursor, hasPrevPage, hasNextPage } =
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

  if (totalCount <= 10) {
    return null;
  }
  return (
    <Pagination
      {...pagination}
      showNumbers={false}
      isPrevDisabled={!hasPrevPage}
      isNextDisabled={!hasNextPage}
      showItemsPerPage={false}
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
  );
};
