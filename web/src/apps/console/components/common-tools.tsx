import { ArrowDown, ArrowUp, ArrowsDownUp, Plus } from '@jengaicons/react';
import { useSearchParams } from '@remix-run/react';
import { useState } from 'react';
import OptionList from '~/components/atoms/option-list';
import Toolbar from '~/components/atoms/toolbar';
import { CommonFilterOptions } from '~/console/components/common-filter';
import Filters, {
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
  const { setQueryParameters } = useQueryParameters();
  const [searchparams] = useSearchParams();
  const page = decodeUrl(searchparams.get('page')) || {};

  const { orderBy = 'updateTime', sortDirection = 'DESC' } = page || {};

  const updateOrder = ({ order, direction }: any) => {
    setQueryParameters({
      page: encodeUrl({
        ...page,
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

export interface ICommonToolsOption {
  name: string;
  type: string;
  search: boolean;
  dataFetcher: (
    s: string
  ) => Promise<{ content: string; value: string | boolean }[]>;
}

export interface IModeProps<T = 'list' | 'grid' | NonNullableString> {
  viewMode?: T;
  setViewMode?: (fn: T) => void;
}

interface ICommonTools extends IModeProps {
  options: ICommonToolsOption[];
  noViewMode?: boolean;
  noSort?: boolean;
}

const CommonTools = ({
  options,
  noViewMode = false,
  noSort = false,
}: ICommonTools) => {
  const [appliedFilters, setAppliedFilters] = useState<IAppliedFilters>({});
  const [sortbyOptionListOpen, setSortybyOptionListOpen] = useState(false);

  useSetAppliedFilters({
    setAppliedFilters,
    types: options,
  });

  return (
    <div className="flex flex-col">
      <div>
        {/* Toolbar for md and up */}
        <div className="hidden md:flex">
          <Toolbar.Root>
            <SearchBox />
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
