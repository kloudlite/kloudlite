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
  Plus,
  PlusFill,
  Search,
  Snowflake,
  SquaresFour,
  Trash,
} from '@jengaicons/react';
import { SubHeader } from '~/components/organisms/sub-header.jsx';
import { Button, IconButton } from '~/components/atoms/button.jsx';
import Toolbar from '~/components/atoms/toolbar';
import OptionList from '~/components/atoms/option-list';
import { Thumbnail } from '~/components/atoms/thumbnail';
import Pagination from '~/components/molecule/pagination';
import { AnimatePresence, motion } from 'framer-motion';
import * as Chips from '~/components/atoms/chips';
import { cn } from '~/components/utils';
import { ChipGroupPaddingTop } from '~/design-system/tailwind-base';
import ResourceList from '../components/resource-list';
import { EmptyState } from '../components/empty-state';
import ScrollArea from '../components/scroll-area';

const ProjectList = [
  {
    name: 'Lobster early',
    id: 'lobster-early-kloudlite-app1',
    cluster: 'Plaxonic',
    path: 'dusty-crossbow.com/projects',
    author: 'Reyan updated the project',
    lastupdated: '3 days ago',
  },
  {
    name: 'Lobster early',
    id: 'lobster-early-kloudlite-app2',
    cluster: 'Plaxonic',
    path: 'dusty-crossbow.com/projects',
    author: 'Reyan updated the project',
    lastupdated: '3 days ago',
  },
  {
    name: 'Lobster early',
    id: 'lobster-early-kloudlite-app3',
    cluster: 'Plaxonic',
    path: 'dusty-crossbow.com/projects',
    author: 'Reyan updated the project',
    lastupdated: '3 days ago',
  },
  {
    name: 'Lobster early',
    id: 'lobster-early-kloudlite-app4',
    cluster: 'Plaxonic',
    path: 'dusty-crossbow.com/projects',
    author: 'Reyan updated the project',
    lastupdated: '3 days ago',
  },
];

const AppliedFilters = [
  {
    id: '0',
    label: 'Active',
    type: Chips.ChipType.REMOVABLE,
    prefix: 'Status:',
  },
  {
    id: '1',
    label: 'Plaxonic',
    type: Chips.ChipType.REMOVABLE,
    prefix: 'Cluster:',
  },
  {
    id: '3',
    label: 'Plaxonic1',
    type: Chips.ChipType.REMOVABLE,
    prefix: 'Cluster:',
  },
  {
    id: '4',
    label: 'Plaxonic2',
    type: Chips.ChipType.REMOVABLE,
    prefix: 'Cluster',
  },
  {
    id: '5',
    label: 'Plaxonic3',
    type: Chips.ChipType.REMOVABLE,
    prefix: 'Cluster:',
  },
  {
    id: '6',
    label: 'Plaxonic4',
    type: Chips.ChipType.REMOVABLE,
    prefix: 'Cluster:',
  },
];

const ProjectToolbar = ({ viewMode, setViewMode }) => {
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

const ProjectFilters = ({ appliedFilters, setAppliedFilters }) => {
  return (
    <AnimatePresence initial={false}>
      {appliedFilters.length > 0 && (
        <motion.div
          className={cn('flex flex-row gap-xl relative')}
          initial={{
            height: 0,
            opacity: 0,
            paddingTop: '0px',
            overflow: 'hidden',
          }}
          animate={{
            height: '46px',
            opacity: 1,
            paddingTop: ChipGroupPaddingTop,
          }}
          exit={{
            height: 0,
            opacity: 0,
            paddingTop: '0px',
            overflow: 'hidden',
          }}
          transition={{
            ease: 'linear',
          }}
          onAnimationStart={(e) => console.log(e)}
        >
          <ScrollArea className="flex-1">
            <Chips.ChipGroup
              onRemove={(c) =>
                setAppliedFilters(appliedFilters.filter((a) => a.id !== c))
              }
            >
              {appliedFilters.map((af) => {
                return <Chips.Chip {...af} key={af.id} item={af} />;
              })}
            </Chips.ChipGroup>
          </ScrollArea>
          {appliedFilters.length > 0 && (
            <div className="flex flex-row items-center justify-center">
              <Button
                content="Clear all"
                variant="primary-plain"
                onClick={() => {
                  setAppliedFilters([]);
                }}
              />
            </div>
          )}
        </motion.div>
      )}
    </AnimatePresence>
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
// mode param is passed from parent element
export const ResourceItem = ({
  mode,
  name,
  id,
  cluster,
  path,
  lastupdated,
  author,
}) => {
  const [openExtra, setOpenExtra] = useState(false);

  const ThumbnailComponent = () => (
    <Thumbnail
      size="small"
      rounded
      src="https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"
    />
  );

  const TitleComponent = () => (
    <>
      <div className="flex flex-row gap-md items-center">
        <div className="headingMd text-text-default">{name}</div>
        <div className="w-lg h-lg bg-icon-primary rounded-full" />
      </div>
      <div className="bodyMd text-text-soft truncate">{id}</div>
    </>
  );

  const ClusterComponent = () => (
    <>
      <div className="bodyMd text-text-strong">{path}</div>
      <div className="bodyMd text-text-strong">{cluster}</div>
    </>
  );

  const AuthorComponent = () => (
    <>
      <div className="bodyMd text-text-strong">{author}</div>
      <div className="bodyMd text-text-soft">{lastupdated}</div>
    </>
  );

  const OptionMenu = () => (
    <ResourceItemExtraOptions open={openExtra} setOpen={setOpenExtra} />
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
            <ThumbnailComponent />
            <div className="flex flex-col gap-sm w-[calc(100%-52px)] md:w-auto">
              {TitleComponent()}
            </div>
          </div>
          {OptionMenu()}
        </div>
        <div className="flex flex-col gap-md items-start">
          {ClusterComponent()}
        </div>
        <div className="flex flex-col items-start">{AuthorComponent()}</div>
      </div>
    );
  };

  const listView = () => (
    <>
      <div className="hidden md:flex flex-row items-center justify-between gap-3xl md:w-full">
        <div className="flex flex-1 flex-row items-center gap-xl">
          <ThumbnailComponent />
          <div className="flex flex-col gap-sm">{TitleComponent()}</div>
        </div>
        {ClusterComponent()}
        <div className="flex flex-col">{AuthorComponent()}</div>
        {OptionMenu()}
      </div>
      {gridView()}
    </>
  );

  if (mode === 'grid') return gridView();
  return listView();
};

// OptionList for various actions
const StatusOptionList = ({ open, setOpen }) => {
  const [statuses, setStatuses] = useState([
    { checked: false, content: 'Active', id: 'active' },
    { checked: false, content: 'Freezed', id: 'freezed' },
    { checked: false, content: 'Archived', id: 'archived' },
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
    </OptionList>
  );
};

const ClusterOptionList = ({ open, setOpen }) => {
  const [clusters, setClusters] = useState([
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
        {clusters.map((cluster) => (
          <OptionList.CheckboxItem
            key={cluster.id}
            checked={cluster.checked}
            onValueChange={(e) =>
              setClusters(
                clusters.map((cltr) => {
                  return cltr.id === cluster.id
                    ? { ...cltr, checked: e }
                    : cltr;
                })
              )
            }
            onSelect={(e) => e.preventDefault()}
          >
            {cluster.content}
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

const ProjectIndex = () => {
  const [projects, _setProjects] = useState(ProjectList);
  const [appliedFilters, setAppliedFilters] = useState(AppliedFilters);
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(15);
  const [totalItems, setTotalItems] = useState(100);
  const [viewMode, setViewMode] = useState('list');

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
          <div className="flex flex-col">
            <ProjectToolbar viewMode={viewMode} setViewMode={setViewMode} />
            <ProjectFilters
              appliedFilters={appliedFilters}
              setAppliedFilters={setAppliedFilters}
            />
          </div>
          <ResourceList mode={viewMode}>
            {projects.map((project) => (
              <ResourceList.ResourceItem key={project.id}>
                <ResourceItem {...project} />
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
        </div>
      )}
      {projects.length === 0 && (
        <div className="pt-3xl">
          <EmptyState
            illustration={
              <svg
                width="226"
                height="227"
                viewBox="0 0 226 227"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <rect y="0.970703" width="226" height="226" fill="#F4F4F5" />
              </svg>
            }
            heading="This is where you’ll manage your projects"
            action={{
              content: 'Add new projects',
              prefix: Plus,
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

export const handler = ({ hi }) => {
  return <div>{hi}</div>;
};

export const loader = () => {
  return {
    hi: 'hello',
  };
};

export default ProjectIndex;
