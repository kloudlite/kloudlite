import {
  ArrowDown,
  ArrowUp,
  ArrowsDownUp,
  CopySimple,
  FunnelSimple,
  Search,
} from '@jengaicons/react';
import { useState } from 'react';

import OptionList from '~/components/atoms/option-list';
import Toolbar from '~/components/atoms/toolbar';
import Pagination from '~/components/molecule/pagination';
import { cn } from '~/components/utils';
import * as Chips from '~/components/atoms/chips';
import ResourceList from '../components/resource-list';
import { dummyData } from '../dummy/data';

const SortbyOptionList = ({ open, setOpen }) => {
  const [sortbyProperty, setSortbyProperty] = useState('updated');
  const [sortbyTime, setSortbyTime] = useState('oldest');
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
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
          value={sortbyProperty}
          onValueChange={setSortbyProperty}
        >
          <OptionList.RadioGroupItem
            value="title"
            onSelect={(e) => e.preventDefault()}
          >
            Container title
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="updated"
            onSelect={(e) => e.preventDefault()}
          >
            Updated
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
        <OptionList.Separator />
        <OptionList.RadioGroup value={sortbyTime} onValueChange={setSortbyTime}>
          <OptionList.RadioGroupItem
            showIndicator={false}
            value="oldest"
            onSelect={(e) => e.preventDefault()}
          >
            <ArrowUp size={16} />
            Oldest first
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="newest"
            showIndicator={false}
            onSelect={(e) => e.preventDefault()}
          >
            <ArrowDown size={16} />
            Newest first
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
      </OptionList.Content>
    </OptionList.Root>
  );
};

const FilterList = ({ open, setOpen }) => {
  const [providers, setProviders] = useState([
    { checked: false, content: 'Linux', id: 'linux' },
    { checked: false, content: 'IBM Z', id: 'ibmz' },
    { checked: false, content: 'Riscv 64', id: 'riscv64' },
    { checked: false, content: 'ARM 64', id: 'arm64' },
    { checked: false, content: 'ARM', id: 'arm' },
  ]);
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Toolbar.Button
          content="Filters"
          variant="basic"
          prefix={<FunnelSimple />}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.TextInput
          value=""
          placeholder="Filter tags"
          prefixIcon={Search}
        />
        {providers.map((provider) => (
          <OptionList.CheckboxItem
            key={provider.id}
            checked={provider.checked}
            onValueChange={(e) =>
              setProviders(
                providers.map((pro) => {
                  return pro.id === provider.id ? { ...pro, checked: e } : pro;
                })
              )
            }
            onSelect={(e) => e.preventDefault()}
          >
            {provider.content}
          </OptionList.CheckboxItem>
        ))}
      </OptionList.Content>
    </OptionList.Root>
  );
};

const CRToolbar = () => {
  const [filterOptionList, setFilterOptionListOpen] = useState(false);
  const [sortbyOptionListOpen, setSortybyOptionListOpen] = useState(false);
  return (
    <div>
      {/* Toolbar for md and up */}
      <div className="hidden md:flex">
        <Toolbar.Root>
          <div className="w-full">
            <Toolbar.TextInput
              value=""
              placeholder="Search"
              prefixIcon={Search}
            />
          </div>
          <FilterList
            open={filterOptionList}
            setOpen={setFilterOptionListOpen}
          />
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
        </Toolbar.Root>
      </div>

      {/* Toolbar for mobile screen */}
      <div className="flex md:hidden">
        <Toolbar.Root>
          <div className="flex-1">
            <Toolbar.TextInput
              value=""
              placeholder="Search"
              prefixIcon={Search}
            />
          </div>
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
        </Toolbar.Root>
      </div>
    </div>
  );
};

// Project resouce item for grid and list mode
// mode param is passed from parent element
export const ResourceItem = ({
  mode = '',
  name,
  id,
  tags = [],
  builds,
  lastupdated,
  author,
}) => {
  const TitleComponent = () => (
    <>
      <div className="flex flex-row gap-md items-center">
        <div className="headingMd text-text-default">{name}</div>
      </div>
      <div className="bodyMd text-text-soft flex flex-row gap-lg items-center">
        <span className="truncate">{id}</span>
        <CopySimple size={12} />
      </div>
    </>
  );

  const BuildComponent = () => (
    <>
      <div className="bodyMd text-text-strong w-[100px]">
        {String(tags.length).padStart(2, '0')} tags
      </div>
      <div className="bodyMd text-text-strong w-[100px]">{builds} builds</div>
    </>
  );

  const AuthorComponent = () => (
    <>
      <div className="bodyMd text-text-strong">{author}</div>
      <div className="bodyMd text-text-soft">{lastupdated}</div>
    </>
  );

  const TagComponent = () => (
    <Chips.ChipGroup>
      {tags.map((tag) => (
        <Chips.Chip
          key={tag}
          label={tag}
          item={{ tag }}
          type={Chips.ChipType.CLICKABLE}
        />
      ))}
    </Chips.ChipGroup>
  );

  const gridView = () => {
    return (
      <div
        className={cn('flex flex-col gap-3xl w-full', {
          'md:hidden': mode === 'list',
        })}
      >
        <div className="flex flex-row items-center justify-between gap-lg w-full">
          <div className="flex flex-row items-center gap-xl w-[calc(100%-44px)] md:w-auto">
            <div className="flex flex-col gap-sm w-[calc(100%-52px)] md:w-auto">
              {TitleComponent()}
            </div>
          </div>
        </div>
        <div className="flex flex-col gap-md items-start">
          {BuildComponent()}
        </div>
        <div className="flex flex-col items-start">{AuthorComponent()}</div>
      </div>
    );
  };

  const listView = () => (
    <>
      <div className="flex flex-col gap-3xl w-full">
        <div className="hidden md:flex flex-row items-center justify-between gap-3xl w-full">
          <div className="flex flex-1 flex-row items-center gap-xl">
            <div className="flex flex-col gap-sm">{TitleComponent()}</div>
          </div>
          {BuildComponent()}
          <div className="flex flex-col w-[200px]">{AuthorComponent()}</div>
        </div>
        {TagComponent()}
      </div>
      {gridView()}
    </>
  );

  if (mode === 'grid') return gridView();
  return listView();
};

const ContainerRegistryGeneral = () => {
  const [crg, _setCrg] = useState(dummyData.containerRegistryGeneralList);
  const [currentPage, _setCurrentPage] = useState(1);
  const [itemsPerPage, _setItemsPerPage] = useState(15);
  const [totalItems, _setTotalItems] = useState(100);
  return (
    <>
      <CRToolbar />
      <ResourceList mode="list">
        {crg.map((cr) => (
          <ResourceList.ResourceItem key={cr.id} textValue={cr.id}>
            <ResourceItem {...cr} />
          </ResourceList.ResourceItem>
        ))}
      </ResourceList>
      <div className="hidden md:flex">
        <Pagination
          currentPage={currentPage}
          itemsPerPage={itemsPerPage}
          totalItems={totalItems}
        />
      </div>
    </>
  );
};

export default ContainerRegistryGeneral;
