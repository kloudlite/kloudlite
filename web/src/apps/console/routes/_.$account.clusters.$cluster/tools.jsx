import Toolbar from '~/components/atoms/toolbar';
import OptionList from '~/components/atoms/option-list';
import { useEffect, useState } from 'react';
import {
  ArrowDown,
  ArrowUp,
  ArrowsDownUp,
  CaretDownFill,
  List,
  Plus,
  Search,
  SquaresFour,
} from '@jengaicons/react';
import { SearchBox } from '~/console/components/search-box';

const Tools = ({ viewMode, setViewMode }) => {
  const [statusOptionListOpen, setStatusOptionListOpen] = useState(false);
  const [typeOptionListOpen, setTypeOptionListOpen] = useState(false);
  const [sortbyOptionListOpen, setSortybyOptionListOpen] = useState(false);
  return (
    <div>
      {/* Toolbar for md and up */}
      <div className="hidden md:flex">
        <Toolbar.Root>
          <SearchBox fields={['metadata.name']} />

          <Toolbar.ButtonGroup.Root value="hello">
            <StatusOptionList
              open={statusOptionListOpen}
              setOpen={setStatusOptionListOpen}
            />
            <TypeOptionList
              open={typeOptionListOpen}
              setOpen={setTypeOptionListOpen}
            />
          </Toolbar.ButtonGroup.Root>
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
          <ViewMode mode={viewMode} onModeChange={setViewMode} />
        </Toolbar.Root>
      </div>

      {/* Toolbar for mobile screen */}
      <div className="flex md:hidden">
        <Toolbar.Root>
          <div className="flex-1">
            <Toolbar.TextInput placeholder="Search" prefixIcon={Search} />
          </div>
          <Toolbar.Button content="Add filters" prefix={Plus} variant="basic" />
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
        </Toolbar.Root>
      </div>
    </div>
  );
};

// OptionList for various actions
const StatusOptionList = ({ open, setOpen }) => {
  const [statuses, setStatuses] = useState([
    { checked: false, content: 'Running', id: 'running' },
    { checked: false, content: 'Stopped', id: 'stopped' },
  ]);
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Toolbar.ButtonGroup.Button
          content="Status"
          variant="basic"
          suffix={CaretDownFill}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        {statuses.map((status) => (
          <OptionList.CheckboxItem
            key={status.id}
            checked={status.checked}
            onValueChange={(e) =>
              setStatuses(
                statuses.map((stat) => {
                  return stat.id === status.id ? { ...stat, checked: e } : stat;
                })
              )
            }
            onSelect={(e) => e.preventDefault()}
          >
            {status.content}
          </OptionList.CheckboxItem>
        ))}
      </OptionList.Content>
    </OptionList.Root>
  );
};

const TypeOptionList = ({ open, setOpen }) => {
  const [regions, setRegions] = useState([
    { checked: false, content: 'On demand', id: 'ondemand' },
    { checked: false, content: 'Spot', id: 'spot' },
    { checked: false, content: 'Reserved', id: 'reserved' },
  ]);
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Toolbar.ButtonGroup.Button
          content="Type"
          variant="basic"
          suffix={CaretDownFill}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        {regions.map((region) => (
          <OptionList.CheckboxItem
            key={region.id}
            checked={region.checked}
            onValueChange={(e) =>
              setRegions(
                regions.map((reg) => {
                  return reg.id === region.id ? { ...reg, checked: e } : reg;
                })
              )
            }
            onSelect={(e) => e.preventDefault()}
          >
            {region.content}
          </OptionList.CheckboxItem>
        ))}
      </OptionList.Content>
    </OptionList.Root>
  );
};

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
              prefix={ArrowsDownUp}
            />
          </div>

          <div className="flex md:hidden">
            <Toolbar.IconButton variant="basic" icon={ArrowsDownUp} />
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
            Node title
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

export default Tools;
