import {
  ArrowDown,
  ArrowUp,
  ArrowsDownUp,
  Plus,
} from '~/console/components/icons';
import { ReactNode, useCallback, useState } from 'react';
import OptionList from '~/components/atoms/option-list';
import Toolbar from '~/components/atoms/toolbar';
import { cn } from '~/components/utils';
import { CommonFilterOptions } from '~/console/components/common-filter';
import Filters, {
  FilterType,
  IAppliedFilters,
  useSetAppliedFilters,
} from '~/console/components/filters';
import { SearchBox } from '~/console/components/search-box';
import ViewMode from '~/console/components/view-mode';
import {
  decodeUrl,
  encodeUrl,
  useQueryParameters,
} from '~/root/lib/client/hooks/use-search';
import { NonNullableString } from '~/root/lib/types/common';

interface ISortbyOptionList {
  open?: boolean;
  setOpen: React.Dispatch<React.SetStateAction<boolean>>;
}

const SortbyOptionList = (_: ISortbyOptionList) => {
  const { setQueryParameters, sparams } = useQueryParameters();

  const page = useCallback(
    () => decodeUrl(sparams.get('page')) || '',
    [sparams]
  )();

  const { orderBy = 'updateTime', sortDirection = 'DESC' } = page || {};

  const updateOrder = ({ order, direction }: any) => {
    const po = { ...page };
    delete po.first;
    delete po.after;
    delete po.last;
    delete po.before;

    setQueryParameters({
      page: encodeUrl({
        ...po,
        orderBy: order,
        sortDirection: direction,
      }),
    });
  };

  return (
    <OptionList.Root>
      <OptionList.Trigger>
        <Toolbar.Button
          content="Sortby"
          variant="basic"
          prefix={<ArrowsDownUp />}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.RadioGroup
          value={orderBy}
          onValueChange={(v) =>
            updateOrder({
              direction: sortDirection,
              order: v,
            })
          }
        >
          <OptionList.RadioGroupItem
            value="metadata.name"
            onClick={(e) => e.preventDefault()}
          >
            Name
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="updateTime"
            onClick={(e) => e.preventDefault()}
          >
            Updated
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
        <OptionList.Separator />
        <OptionList.RadioGroup
          value={sortDirection}
          onValueChange={(v) =>
            updateOrder({
              order: orderBy,
              direction: v,
            })
          }
        >
          <OptionList.RadioGroupItem
            showIndicator={false}
            value="ASC"
            onClick={(e) => e.preventDefault()}
          >
            <ArrowUp size={16} />
            {orderBy === 'updateTime' ? 'Oldest' : 'Ascending'}
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="DESC"
            showIndicator={false}
            onClick={(e) => e.preventDefault()}
          >
            <ArrowDown size={16} />
            {orderBy === 'updateTime' ? 'Newest' : 'Descending'}
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
      </OptionList.Content>
    </OptionList.Root>
  );
};

export interface IModeProps<T = 'list' | 'grid' | NonNullableString> {
  // eslint-disable-next-line react/no-unused-prop-types
  viewMode?: T;
  // eslint-disable-next-line react/no-unused-prop-types
  setViewMode?: (fn: T) => void;
}

interface ICommonTools extends IModeProps {
  options: FilterType[];
  noViewMode?: boolean;
  noSort?: boolean;
  commonToolPrefix?: ReactNode;
}

const CommonTools = ({
  options,
  noViewMode = false,
  noSort = false,
  commonToolPrefix,
}: ICommonTools) => {
  const [appliedFilters, setAppliedFilters] = useState<IAppliedFilters>({});
  const [sortbyOptionListOpen, setSortybyOptionListOpen] = useState(false);

  // eslint-disable-next-line no-param-reassign
  noViewMode = true;

  useSetAppliedFilters({
    setAppliedFilters,
    types: options,
  });

  return (
    <div className={cn('flex flex-col bg-surface-basic-subdued pb-6xl')}>
      <div>
        {/* Toolbar for md and up */}
        <div className="hidden md:flex">
          <Toolbar.Root>
            <SearchBox />
            {commonToolPrefix}
            <CommonFilterOptions options={options} />
            {!noSort && (
              <SortbyOptionList
                open={sortbyOptionListOpen}
                setOpen={setSortybyOptionListOpen}
              />
            )}
            {!noViewMode && <ViewMode />}
          </Toolbar.Root>
        </div>

        {/* Toolbar for mobile screen */}
        <div className="flex md:hidden">
          <Toolbar.Root>
            <div className="flex-1">
              <SearchBox />
            </div>
            <Toolbar.Button
              content="Add filters"
              prefix={<Plus />}
              variant="basic"
            />
            <SortbyOptionList
              open={sortbyOptionListOpen}
              setOpen={setSortybyOptionListOpen}
            />
          </Toolbar.Root>
        </div>
      </div>

      <Filters appliedFilters={appliedFilters} />
    </div>
  );
};

export default CommonTools;
