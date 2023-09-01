import { NonNullableString } from '~/root/lib/types/common';

interface InewPagination {
  orderBy?: 'metadata.name' | 'updateTime' | NonNullableString;
  sortBy?: 'ASC' | 'DESC' | NonNullableString;
  last?: number;
  first?: number;
  before?: string;
  after?: string;
}

export const newPagination = ({
  orderBy,
  sortBy,
  last,
  first,
  before,
  after,
}: InewPagination) => {
  return {
    ...{
      orderBy,
      sortBy,
      last,
      first,
      before,
      after,
    },
  };
};
