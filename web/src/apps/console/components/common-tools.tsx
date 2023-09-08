import Toolbar from '~/components/atoms/toolbar';
import OptionList from '~/components/atoms/option-list';
import { useState } from 'react';
import { ArrowDown, ArrowUp, ArrowsDownUp, Plus } from '@jengaicons/react';
import { SearchBox } from '~/console/components/search-box';
import ViewMode from '~/console/components/view-mode';
import { CommonFilterOptions } from '~/console/components/common-filter';
import Filters, {
  IAppliedFilters,
  useSetAppliedFilters,
} from '~/console/components/filters';
import {
  decodeUrl,
  encodeUrl,
  useQueryParameters,
} from '~/root/lib/client/hooks/use-search';
import { useSearchParams } from '@remix-run/react';
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
        <div>
          <div className="hidden md:flex">
            <Toolbar.Button
              content="Sortby"
              variant="basic"
              prefix={<ArrowsDownUp />}
            />
          </div>

          <div className="flex md:hidden">
            <Toolbar.IconButton variant="basic" icon={<ArrowsDownUp />} />
          </div>
        </div>
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
            onSelect={(e) => e.preventDefault()}
          >
            Name
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="updateTime"
            onSelect={(e) => e.preventDefault()}
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
            onSelect={(e) => e.preventDefault()}
          >
            <ArrowUp size={16} />
            {orderBy === 'updateTime' ? 'Oldest' : 'Ascending'}
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="DESC"
            showIndicator={false}
            onSelect={(e) => e.preventDefault()}
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
}

const CommonTools = ({
  viewMode = 'list',
  setViewMode = (_) => _,
  options,
  noViewMode = false,
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
            <SortbyOptionList
              open={sortbyOptionListOpen}
              setOpen={setSortybyOptionListOpen}
            />
            {!noViewMode && (
              <ViewMode mode={viewMode} onModeChange={setViewMode} />
            )}
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
