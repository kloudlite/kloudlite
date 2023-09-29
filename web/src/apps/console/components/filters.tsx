import { useLocation, useSearchParams } from '@remix-run/react';
import { AnimatePresence, motion } from 'framer-motion';
import { ReactElement, useEffect, useState } from 'react';
import useSWR from 'swr';
import { Button } from '~/components/atoms/button';
import Chips from '~/components/atoms/chips';
import ScrollArea from '~/components/atoms/scroll-area';
import { cn } from '~/components/utils';
import { ChipGroupPaddingTop } from '~/design-system/tailwind-base';
import {
  IQueryParams,
  decodeUrl,
  encodeUrl,
  useQueryParameters,
} from '~/root/lib/client/hooks/use-search';

interface IremoveFilter {
  type: string;
  value: string;
  searchParams: URLSearchParams;
  setQueryParameters: (v: IQueryParams) => void;
}

const removeFilter = ({
  type,
  value,
  searchParams,
  setQueryParameters,
}: IremoveFilter) => {
  const search = decodeUrl(searchParams.get('search'));
  const array = search?.[type]?.array || [];

  let nArray = [];
  nArray = array.filter((_v: any) => _v !== value);

  if (nArray.length === 0) {
    if (search[type]) {
      delete search[type];
    }

    setQueryParameters({
      search: encodeUrl(search),
    });
  } else {
    setQueryParameters({
      search: encodeUrl({
        ...search,
        [type]: {
          matchType: 'array',
          array: nArray,
        },
      }),
    });
  }
};

export type IdataFetcher = (
  s: string
) => Promise<{ content: string; value: string | boolean; type?: string }[]>;

export interface FilterType {
  name: string;
  type: string;
  search: boolean;
  dataFetcher: IdataFetcher;
}

export interface IAppliedFilters {
  [name: string]: { type: string; array: string[]; dataFetcher: IdataFetcher };
}

interface IuseSetAppliedFilters {
  setAppliedFilters: React.Dispatch<React.SetStateAction<IAppliedFilters>>;
  types: FilterType[];
}

export const useSetAppliedFilters = ({
  setAppliedFilters,
  types = [],
}: IuseSetAppliedFilters) => {
  const [searchParams] = useSearchParams();
  useEffect(() => {
    const filters = decodeUrl(searchParams.get('search')) || {};
    setAppliedFilters((s) => {
      return {
        ...s,
        ...types.reduce((acc, { name, type, dataFetcher }) => {
          return {
            ...acc,
            [name]: {
              type,
              array: filters[type]?.array || [],
              dataFetcher,
            },
          };
        }, {}),
      };
    });
  }, [searchParams]);
};

const FilterLabel = ({
  dataFetcher,
  item,
  type,
}: {
  dataFetcher: IdataFetcher;
  item: string;
  type: string;
}) => {
  const location = useLocation();

  const { data } = useSWR(`${location.pathname}-${type}`, dataFetcher);
  return <span>{data?.find((v) => v.value === item)?.content || item}</span>;
};

const Filters = ({ appliedFilters }: { appliedFilters: IAppliedFilters }) => {
  const [chipsData, setChipsData] = useState<ReactElement[]>([]);

  const [searchParams] = useSearchParams();
  const { setQueryParameters } = useQueryParameters();

  useEffect(() => {
    setChipsData(
      Object.entries(appliedFilters).reduce<ReactElement[]>(
        (acc, [key, { array, type, dataFetcher }]) => {
          return [
            ...acc,
            ...array.map((item) => (
              <Chips.Chip
                key={`${type}:${item}`}
                item={{ type, value: item }}
                label={<FilterLabel {...{ dataFetcher, item, type }} />}
                prefix={`${key}:`}
                type="REMOVABLE"
              />
            )),
          ];
        },
        []
      )
    );
  }, [appliedFilters]);
  return (
    <AnimatePresence initial={false}>
      {chipsData.length > 0 && (
        <motion.div
          className={cn('flex flex-row gap-xl relative -ml-[3px] pl-[3px]')}
          initial={{
            height: 0,
            opacity: 0,
            paddingTop: '0px',
            overflow: 'hidden',
          }}
          animate={{
            height: '46px',
            opacity: 1,
            paddingTop: ChipGroupPaddingTop,
          }}
          exit={{
            height: 0,
            opacity: 0,
            paddingTop: '0px',
            overflow: 'hidden',
          }}
          transition={{
            ease: 'linear',
          }}
          // onAnimationStart={(e) => console.log(e)}
        >
          <ScrollArea className="flex-1">
            <Chips.ChipGroup
              onRemove={({ type, value }) => {
                removeFilter({
                  type,
                  value,
                  searchParams,
                  setQueryParameters,
                });
              }}
            >
              {chipsData}
            </Chips.ChipGroup>
          </ScrollArea>
          {chipsData.length && (
            <div className="flex flex-row items-center justify-center">
              <Button
                content="Clear all"
                variant="primary-plain"
                onClick={() => {
                  setQueryParameters({
                    search: encodeUrl({}),
                  });
                }}
              />
            </div>
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default Filters;
