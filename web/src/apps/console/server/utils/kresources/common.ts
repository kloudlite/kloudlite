import { MapType } from '~/root/lib/types/common';

export interface MetadataProps {
  name: string;
  namespace?: string;
  labels: MapType;
  annotations: MapType;
}

export interface IPagination {
  pageInfo: {
    startCursor: string;
    endCursor: string;
    hasPreviousPage: boolean;
    hasNextPage: boolean;
  };
  totalCount: number;
}
