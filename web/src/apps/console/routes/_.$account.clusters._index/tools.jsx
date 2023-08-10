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
        <Toolbar>
          <SearchBox fields={['metadata.name']} />
          <Toolbar.ButtonGroup>
            <ProviderOptionList
              open={clusterOptionListOpen}
              setOpen={setClusterOptionListOpen}
            />
            <RegionOptionList
              open={clusterOptionListOpen}
              setOpen={setClusterOptionListOpen}
            />
            <StatusOptionList
              open={statusOptionListOpen}
              setOpen={setStatusOptionListOpen}
            />
          </Toolbar.ButtonGroup>
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
          <ViewMode mode={viewMode} onModeChange={setViewMode} />
        </Toolbar>
      </div>

      {/* Toolbar for mobile screen */}
      <div className="flex md:hidden">
        <Toolbar>
          <div className="flex-1">
            <SearchBox fields={['metadata.name']} />
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

// OptionList for various actions
const StatusOptionList = ({ open, setOpen }) => {
  const [data, setData] = useState([
    { checked: false, content: 'Connected', id: 'connected' },
    { checked: false, content: 'Disconnected', id: 'disconnected' },
    { checked: false, content: 'Waiting to connect', id: 'waitingtoconnect' },
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
  const [data, setData] = useState([
    { checked: false, content: 'AWS', id: 'aws' },
    { checked: false, content: 'Azure', id: 'azure' },
    { checked: false, content: 'CloudStack', id: 'cloudstack' },
    { checked: false, content: 'Digital Ocean', id: 'digitalocean' },
  ]);
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
          <OptionList.CheckboxItem
            key={d.id}
            checked={d.checked}
            onValueChange={(e) =>
              setData(
                data.map((pro) => {
                  return pro.id === d.id ? { ...pro, checked: e } : pro;
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

const RegionOptionList = ({ open, setOpen }) => {
  const [data, setData] = useState([
    { checked: false, content: 'India', id: 'india' },
    { checked: false, content: 'Asia Pacific', id: 'asiapacific' },
    { checked: false, content: 'Europe', id: 'europe' },
    { checked: false, content: 'Middle East/Africa', id: 'middleeast/africa' },
  ]);
  return (
    <OptionList open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Toolbar.ButtonGroup.Button
          content="Region"
          variant="basic"
          suffix={CaretDownFill}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.TextInput placeholder="Filter region" prefixIcon={Search} />
        {data.map((d) => (
          <OptionList.CheckboxItem
            key={d.id}
            checked={d.checked}
            onValueChange={(e) =>
              setData(
                data.map((reg) => {
                  return reg.id === d.id ? { ...reg, checked: e } : reg;
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
            Cluster title
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
