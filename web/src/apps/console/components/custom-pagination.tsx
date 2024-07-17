import { useSearchParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import Pagination from '~/components/molecule/pagination';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { useLog } from '~/root/lib/client/hooks/use-log';
import {
  decodeUrl,
  encodeUrl,
  useQueryParameters,
} from '~/root/lib/client/hooks/use-search';

export const CustomPagination = ({ pagination }: { pagination: any }) => {
  const { startCursor, endCursor, hasPrevPage, hasNextPage } =
    pagination?.pageInfo || {};

  const { totalCount } = pagination || {};

  const getCursor = (c: any) => {
    const cLength = c?.edges?.length || 0;
    return {
      sCursor: c?.edges?.[0]?.cursor,
      ecursor: c?.edges?.[cLength - 1]?.cursor,
    };
  };
  const [cursor, setCursor] = useState(() => getCursor(pagination));

  useEffect(() => {
    setCursor(getCursor(pagination));
  }, [pagination]);

  useLog(cursor);

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
        if (cursor.ecursor) {
          setPage(newPagination({ first: 10, after: cursor.ecursor }));
        } else {
          setPage(newPagination({ first: 10 }));
        }
      }}
      onClickPrev={() => {
        if (cursor.sCursor) {
          setPage(newPagination({ last: 10, before: cursor.sCursor }));
        } else {
          setPage(newPagination({ last: 10 }));
        }
      }}
    />
  );
};
