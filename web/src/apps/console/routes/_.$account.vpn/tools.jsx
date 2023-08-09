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

const Tools = () => {
  const [clusterOptionListOpen, setClusterOptionList] = useState(false);
  const [authorOptionList, setAuthorOptionList] = useState(false);
  const [sortbyOptionListOpen, setSortybyOptionListOpen] = useState(false);
  return (
    <div>
      {/* Toolbar for md and up */}
      <div className="hidden md:flex">
        <Toolbar>
          <div className="w-full">
            <Toolbar.TextInput placeholder="Search" prefixIcon={Search} />
          </div>
          <Toolbar.ButtonGroup value="hello">
            <ClusterOptionList
              open={clusterOptionListOpen}
              setOpen={setClusterOptionList}
            />
            <AuthorOptionList
              open={authorOptionList}
              setOpen={setAuthorOptionList}
            />
          </Toolbar.ButtonGroup>
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
          {/* <ViewToggle mode={viewMode} onModeChange={setViewMode} /> */}
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

// OptionList for various actions
const ClusterOptionList = ({ open, setOpen }) => {
  const [data, setData] = useState([
    { checked: false, content: 'Plaxonic', id: 'plaxonic' },
    { checked: false, content: 'Hyades', id: 'hyades' },
  ]);
  return (
    <OptionList open={open} onOpenChange={setOpen}>
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
    </OptionList>
  );
};

const AuthorOptionList = ({ open, setOpen }) => {
  const [data, setData] = useState([
    { checked: false, content: 'Noah', id: 'Noah' },
    { checked: false, content: 'Kenely', id: 'Kenely' },
    { checked: false, content: 'Emma', id: 'Emma' },
    { checked: false, content: 'Sofia', id: 'Sofia' },
    { checked: false, content: 'Scarlet', id: 'Scarlet' },
  ]);
  return (
    <OptionList open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Toolbar.ButtonGroup.Button
          content="Author"
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
            VPN title
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
