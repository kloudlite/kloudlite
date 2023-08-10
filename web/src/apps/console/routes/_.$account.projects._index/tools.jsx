import Toolbar from '~/components/atoms/toolbar';
import OptionList from '~/components/atoms/option-list';
import { useState } from 'react';
import {
  ArrowDown,
  ArrowUp,
  ArrowsDownUp,
  CaretDownFill,
  Plus,
  Search,
} from '@jengaicons/react';
import { SearchBox } from '~/console/components/search-box';
import ViewMode from '~/console/components/view-mode';

const Tools = ({ viewMode, setViewMode }) => {
  const [statusOptionListOpen, setStatusOptionListOpen] = useState(false);
  const [clusterOptionListOpen, setClusterOptionListOpen] = useState(false);
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
            <ClusterOptionList
              open={clusterOptionListOpen}
              setOpen={setClusterOptionListOpen}
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
            <SearchBox fields={['metadata.name']} />
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
  const [data, setData] = useState([
    { checked: false, content: 'Active', id: 'active' },
    { checked: false, content: 'Freezed', id: 'freezed' },
    { checked: false, content: 'Archived', id: 'archived' },
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
        {data.map((d) => (
          <OptionList.CheckboxItem
            key={d.id}
            checked={d.checked}
            onValueChange={(e) =>
              setData(
                data.map((stat) => {
                  return stat.id === d.id ? { ...stat, checked: e } : stat;
                })
              )
            }
            onSelect={(e) => e.preventDefault()}
          >
            {d.content}
          </OptionList.CheckboxItem>
        ))}
      </OptionList.Content>
    </OptionList.Root>
  );
};

const ClusterOptionList = ({ open, setOpen }) => {
  const [data, setData] = useState([
    { checked: false, content: 'Plaxonic', id: 'plaxonic' },
    { checked: false, content: 'Hyades', id: 'hyades' },
  ]);
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Toolbar.ButtonGroup.Button
          content="Cluster"
          variant="basic"
          suffix={CaretDownFill}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.TextInput
          placeholder="Filter cluster"
          prefixIcon={Search}
        />
        {data.map((d) => (
          <OptionList.CheckboxItem
            key={d.id}
            checked={d.checked}
            onValueChange={(e) =>
              setData(
                data.map((cltr) => {
                  return cltr.id === d.id ? { ...cltr, checked: e } : cltr;
                })
              )
            }
            onSelect={(e) => e.preventDefault()}
          >
            {d.content}
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
            Project title
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
