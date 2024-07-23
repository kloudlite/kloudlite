import { useCallback } from 'react';
import Pagination from '~/components/molecule/pagination';
import {
  decodeUrl,
  encodeUrl,
  useQueryParameters,
} from '~/root/lib/client/hooks/use-search';

export const CustomPagination = ({ pagination }: { pagination: any }) => {
  const { startCursor, endCursor, hasPrevPage, hasNextPage } =
    pagination?.pageInfo || {};

  const { totalCount } = pagination || {};

  const { setQueryParameters, sparams } = useQueryParameters();
  const page = useCallback(
    () => decodeUrl(sparams.get('page')) || '',
    [sparams]
  )();

  const updatePage = (p: { [key: string]: string | number }) => {
    const po = { ...page };
    if (p.first || p.after) {
      delete po.last;
      delete po.before;
    }

    if (p.last || p.before) {
      delete po.first;
      delete po.after;
    }

    setQueryParameters({ page: encodeUrl({ ...po, ...p }) }, false);
  };

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
          updatePage({
            first: 10,
            after: endCursor,
          });
        } else {
          updatePage({
            ...{ first: 10 },
          });
        }
      }}
      onClickPrev={() => {
        if (startCursor) {
          updatePage({
            last: 10,
            before: startCursor,
          });
        } else {
          updatePage({
            ...{ last: 10 },
          });
        }
      }}
    />
  );
};
