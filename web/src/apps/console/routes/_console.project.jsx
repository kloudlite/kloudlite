import { useEffect, useState } from 'react';
import { Link } from '@remix-run/react';
import {
  Archive,
  ArrowDown,
  ArrowUp,
  ArrowsDownUp,
  CaretDownFill,
  DotsThreeVerticalFill,
  List,
  PlusFill,
  Search,
  Snowflake,
  SquaresFour,
  Trash,
} from '@jengaicons/react';
import { SubHeader } from '~/components/organisms/sub-header.jsx';
import { Button, IconButton } from '~/components/atoms/button.jsx';
import { EmptyState } from '~/components/molecule/empty-state.jsx';
import Toolbar from '~/components/atoms/toolbar';
import OptionList from '~/components/atoms/option-list';
import ChipGroup from '~/components/atoms/chip-group';
import { Thumbnail } from '~/components/atoms/thumbnail';
import ResourceList from '../components/resource-list';

const AppliedFilters = [
  {
    id: '0',
    label: 'Active',
    type: ChipGroup.ChipType.REMOVABLE,
    prefix: 'Status',
  },
  {
    id: '1',
    label: 'Plaxonic',
    type: ChipGroup.ChipType.REMOVABLE,
    prefix: 'Cluster',
  },
  {
    id: '3',
    label: 'Plaxonic1',
    type: ChipGroup.ChipType.REMOVABLE,
    prefix: 'Cluster',
  },
  {
    id: '4',
    label: 'Plaxonic2',
    type: ChipGroup.ChipType.REMOVABLE,
    prefix: 'Cluster',
  },
  {
    id: '5',
    label: 'Plaxonic3',
    type: ChipGroup.ChipType.REMOVABLE,
    prefix: 'Cluster',
  },
  {
    id: '6',
    label: 'Plaxonic4',
    type: ChipGroup.ChipType.REMOVABLE,
    prefix: 'Cluster',
  },
];

const Project = () => {
  const [projects, _setProjects] = useState([1]);
  const [statusOptionListOpen, setStatusOptionListOpen] = useState(false);
  const [clusterOptionListOpen, setClusterOptionListOpen] = useState(false);
  const [sortbyOptionListOpen, setSortybyOptionListOpen] = useState(false);
  const [viewMode, setViewMode] = useState('list');
  const [appliedFilters, setAppliedFilters] = useState(AppliedFilters);

  return (
    <>
      <SubHeader
        title="Projects"
        actions={
          projects.length !== 0 && (
            <Button
              variant="primary"
              content="Add new"
              prefix={PlusFill}
              href="/new-project"
              LinkComponent={Link}
            />
          )
        }
      />
      {projects.length > 0 && (
        <div className="pt-3xl flex flex-col gap-6xl">
          <div className="flex flex-col gap-xl">
            <Toolbar>
              <Toolbar.TextInput
                placeholder="Search"
                prefixIcon={Search}
                className="w-full"
              />
              <Toolbar.ButtonGroup value="hello">
                <StatusOptionList
                  open={statusOptionListOpen}
                  setOpen={setStatusOptionListOpen}
                />
                <ClusterOptionList
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
            <div className="flex flex-row gap-xl">
              <ChipGroup
                onRemove={(c) =>
                  setAppliedFilters(appliedFilters.filter((a) => a.id !== c))
                }
              >
                {appliedFilters.map((af) => {
                  return <ChipGroup.Chip {...af} key={af.id} />;
                })}
              </ChipGroup>
              {appliedFilters.length > 0 && (
                <Button
                  content="Clear all"
                  variant="primary-plain"
                  onClick={() => {
                    setAppliedFilters([]);
                  }}
                />
              )}
            </div>
          </div>
          <ResourceList mode={viewMode}>
            <ResourceList.ResourceItem>
              <ResourceItem mode={viewMode} />
            </ResourceList.ResourceItem>
          </ResourceList>
        </div>
      )}
      {projects.length === 0 && (
        <div className="pt-5">
          <EmptyState
            heading="This is where you’ll manage your projects"
            action={{
              title: 'Create Project',
              LinkComponent: Link,
              href: '/new-project',
            }}
          >
            <p>You can create a new project and manage the listed project.</p>
          </EmptyState>
        </div>
      )}
    </>
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

// Project resouce item for grid and list mode
export const ResourceItem = ({ mode = 'list' }) => {
  const [openExtra, setOpenExtra] = useState(false);
  if (mode === 'grid')
    return (
      <>
        <div className="flex flex-row items-center justify-between gap-lg">
          <div className="flex flex-row items-center gap-xl">
            <Thumbnail
              size="small"
              rounded
              src="https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"
            />
            <div className="flex flex-col gap-sm">
              <div className="flex flex-row gap-md items-center">
                <div className="headingMd text-text-default">Lobster early</div>
                <div className="w-lg h-lg bg-icon-primary rounded-full" />
              </div>
              <div className="bodyMd text-text-soft">
                lobster-early-kloudlite-app
              </div>
            </div>
          </div>
          <ResourceItemExtraOptions open={openExtra} setOpen={setOpenExtra} />
        </div>
        <div className="flex flex-col gap-md items-start">
          <div className="bodyMd text-text-strong">
            dusty-crossbow.com/projects
          </div>
          <div className="bodyMd text-text-strong">Plaxonic</div>
        </div>
        <div className="flex flex-col items-start">
          <div className="bodyMd text-text-strong">
            Reyan updated the project
          </div>
          <div className="bodyMd text-text-soft">3 days ago</div>
        </div>
      </>
    );
  return (
    <>
      <div className="flex flex-row items-center gap-xl">
        <Thumbnail
          size="small"
          rounded
          src="https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"
        />
        <div className="flex flex-col gap-sm">
          <div className="flex flex-row gap-md items-center">
            <div className="headingMd text-text-default">Lobster early</div>
            <div className="w-lg h-lg bg-icon-primary rounded-full" />
          </div>
          <div className="bodyMd text-text-soft">
            lobster-early-kloudlite-app
          </div>
        </div>
      </div>
      <div className="bodyMd text-text-strong">dusty-crossbow.com/projects</div>
      <div className="bodyMd text-text-strong">Plaxonic</div>
      <div className="flex flex-col">
        <div className="bodyMd text-text-strong">Reyan updated the project</div>
        <div className="bodyMd text-text-soft">3 days ago</div>
      </div>
      <ResourceItemExtraOptions open={openExtra} setOpen={setOpenExtra} />
    </>
  );
};

// OptionList for various actions
const StatusOptionList = ({ open, setOpen }) => {
  const [active, setActive] = useState(false);
  const [freezed, setFreezed] = useState(false);
  const [archived, setArchived] = useState(false);
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
        <OptionList.CheckboxItem
          checked={active}
          onValueChange={setActive}
          onSelect={(e) => e.preventDefault()}
        >
          Active
        </OptionList.CheckboxItem>
        <OptionList.CheckboxItem
          checked={freezed}
          onValueChange={setFreezed}
          onSelect={(e) => e.preventDefault()}
        >
          Freezed
        </OptionList.CheckboxItem>

        <OptionList.CheckboxItem
          checked={archived}
          onValueChange={setArchived}
          onSelect={(e) => e.preventDefault()}
        >
          Archived
        </OptionList.CheckboxItem>
      </OptionList.Content>
    </OptionList>
  );
};

const ClusterOptionList = ({ open, setOpen }) => {
  const [plaxonic, setPlaxonic] = useState(false);
  const [hyades, setHyades] = useState(false);
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
        <OptionList.CheckboxItem
          checked={plaxonic}
          onValueChange={setPlaxonic}
          onSelect={(e) => e.preventDefault()}
        >
          Plaxonic
        </OptionList.CheckboxItem>
        <OptionList.CheckboxItem
          checked={hyades}
          onValueChange={setHyades}
          onSelect={(e) => e.preventDefault()}
        >
          Hyades
        </OptionList.CheckboxItem>
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
        <Toolbar.Button
          content="Sortby"
          variant="basic"
          prefix={ArrowsDownUp}
        />
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
            Product title
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

const ResourceItemExtraOptions = ({ open, setOpen }) => {
  return (
    <OptionList open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <IconButton
          variant="plain"
          icon={DotsThreeVerticalFill}
          selected={open}
          onClick={(e) => {
            e.stopPropagation();
          }}
          onMouseDown={(e) => {
            e.stopPropagation();
          }}
          onPointerDown={(e) => {
            e.stopPropagation();
          }}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.Item>
          <Archive size={16} />
          <span>Archive</span>
        </OptionList.Item>
        <OptionList.Item>
          <Snowflake size={16} />
          <span>Freezed</span>
        </OptionList.Item>
        <OptionList.Separator />
        <OptionList.Item className="!text-text-critical">
          <Trash size={16} />
          <span>Delete</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList>
  );
};

export default Project;
