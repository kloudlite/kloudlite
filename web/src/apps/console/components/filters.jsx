import { AnimatePresence, motion } from 'framer-motion';
import { cn } from '~/components/utils';
import ScrollArea from '~/console/components/scroll-area';
import * as Chips from '~/components/atoms/chips';
import { ChipGroupPaddingTop } from '~/design-system/tailwind-base';
import { Button } from '~/components/atoms/button';
import { useEffect, useState } from 'react';
import {
  decodeUrl,
  encodeUrl,
  useQueryParameters,
} from '~/root/lib/client/hooks/use-search';
import { useSearchParams } from '@remix-run/react';

const removeFilter = ({ type, value, searchParams, setQueryParameters }) => {
  const search = decodeUrl(searchParams.get('search'));
  const array = search?.[type]?.array || [];

  let nArray = [];
  nArray = array.filter((_v) => _v !== value);

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

export const useSetAppliedFilters = ({ setAppliedFilters, types = [] }) => {
  const [searchParams] = useSearchParams();
  useEffect(() => {
    const filters = decodeUrl(searchParams.get('search')) || {};
    setAppliedFilters((s) => {
      return {
        ...s,
        ...types.reduce((acc, { name, type }) => {
          return {
            ...acc,
            [name]: {
              type,
              array: filters[type]?.array || [],
            },
          };
        }, {}),
      };
    });
  }, [searchParams]);
};

const Filters = ({
  appliedFilters,
  // onClose = (_) => _,
  // onClearAll = (_) => _,
}) => {
  const [chipsData, setChipsData] = useState([]);

  const [searchParams] = useSearchParams();
  const { setQueryParameters } = useQueryParameters();

  useEffect(() => {
    setChipsData(
      Object.entries(appliedFilters).reduce(
        // @ts-ignore
        (acc, [key, { array, type }]) => {
          return [
            ...acc,
            ...array.map((item) => (
              <Chips.Chip
                key={`${type}:${item}`}
                item={{ type, value: item }}
                label={`${item}`}
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
          className={cn('flex flex-row gap-xl relative')}
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
