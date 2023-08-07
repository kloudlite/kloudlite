import Toolbar from '~/components/atoms/toolbar';
import OptionList from '~/components/atoms/option-list';
import { useEffect, useState } from 'react';
import {
  ArrowDown,
  ArrowUp,
  ArrowsDownUp,
  CaretDownFill,
  CopySimple,
  List,
  Plus,
  Search,
  SquaresFour,
} from '@jengaicons/react';
import { dummyData } from '~/console/dummy/data';

const Tools = ({ viewMode, setViewMode }) => {
  const [statusOptionListOpen, setStatusOptionListOpen] = useState(false);
  const [clusterOptionListOpen, setClusterOptionListOpen] = useState(false);
  const [sortbyOptionListOpen, setSortybyOptionListOpen] = useState(false);
  return (
    <div>
      {/* Toolbar for md and up */}
      <div className="hidden md:flex">
        <Toolbar>
          <div className="w-full">
            <Toolbar.TextInput placeholder="Search" prefixIcon={Search} />
          </div>
          <Toolbar.ButtonGroup>
            <StatusOptionList
              open={statusOptionListOpen}
              setOpen={setStatusOptionListOpen}
            />
            <ProviderOptionList
              open={clusterOptionListOpen}
              setOpen={setClusterOptionListOpen}
            />
          </Toolbar.ButtonGroup>
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
          <ViewToggle mode={viewMode} onModeChange={setViewMode} />
        </Toolbar>
      </div>

      {/* Toolbar for mobile screen */}
      <div className="flex md:hidden">
        <Toolbar>
          <div className="flex-1">
            <Toolbar.TextInput placeholder="Search" prefixIcon={Search} />
          </div>
          <Toolbar.Button content="Add filters" prefix={Plus} variant="basic" />
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
        </Toolbar>
      </div>
    </div>
  );
};

// Button for toggling between grid and list view
const ViewToggle = ({ mode, onModeChange }) => {
  const [m, setM] = useState(mode);
  useEffect(() => {
    if (onModeChange) onModeChange(m);
  }, [m]);
  return (
    <Toolbar.ButtonGroup value={m} onValueChange={setM} selectable>
      <Toolbar.ButtonGroup.IconButton icon={List} value="list" />
      <Toolbar.ButtonGroup.IconButton icon={SquaresFour} value="grid" />
    </Toolbar.ButtonGroup>
  );
};

// OptionList for various actions
// OptionList for various actions
const StatusOptionList = ({ open, setOpen }) => {
  const [data, setData] = useState([
    { checked: false, content: 'Verified', id: 'verified' },
    { checked: false, content: 'Un-Verified', id: 'unverified' },
  ]);
  return (
    <OptionList open={open} onOpenChange={setOpen}>
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
    </OptionList>
  );
};

const ProviderOptionList = ({ open, setOpen }) => {
  const [data, setData] = useState(dummyData.providers);
  return (
    <OptionList open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Toolbar.ButtonGroup.Button
          content="Provider"
          variant="basic"
          suffix={CaretDownFill}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.TextInput
          placeholder="Filter provider"
          prefixIcon={Search}
        />
        {data.map((d) => (
          <OptionList.Item key={d.id} onSelect={(e) => e.preventDefault()}>
            <div className="flex flex-row gap-xl">
              <CopySimple size={16} />
              {d.content}
            </div>
          </OptionList.Item>
        ))}
      </OptionList.Content>
    </OptionList>
  );
};

const SortbyOptionList = ({ open, setOpen }) => {
  const [sortbyProperty, setSortbyProperty] = useState('updated');
  const [sortbyTime, setSortbyTime] = useState('oldest');
  return (
    <OptionList open={open} onOpenChange={setOpen}>
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
            Provider name
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
    </OptionList>
  );
};

export default Tools;
