import { useEffect, useState } from 'react';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import {
  ArrowDown,
  ArrowUp,
  ArrowsDownUp,
  CaretDownFill,
  DotsThreeVerticalFill,
  List,
  Plus,
  PlusFill,
  Search,
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
import logger from '~/root/lib/client/helpers/log';
import ResourceList from '../components/resource-list';
import { EmptyState } from '../components/empty-state';
import ScrollArea from '../components/scroll-area';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { dummyData } from '../dummy/data';

const ClusterToolbar = ({ viewMode, setViewMode }) => {
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

const ClusterFilters = ({ appliedFilters, setAppliedFilters }) => {
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
  providerRegion,
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
    <div className="bodyMd text-text-strong">{providerRegion}</div>
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
      <div className="hidden md:flex flex-row items-center justify-between gap-3xl w-full">
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

const ProviderOptionList = ({ open, setOpen }) => {
  const [providers, setProviders] = useState([
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
    </OptionList>
  );
};

const RegionOptionList = ({ open, setOpen }) => {
  const [regions, setRegions] = useState([
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
        <OptionList.Item className="!text-text-critical">
          <Trash size={16} />
          <span>Delete</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList>
  );
};

const ClustersIndex = () => {
  const [appliedFilters, setAppliedFilters] = useState(
    dummyData.appliedFilters
  );
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(15);
  const [totalItems, setTotalItems] = useState(100);
  const [viewMode, setViewMode] = useState('list');

  const { clustersData } = useLoaderData();
  const [clusters, _setClusters] = useState(
    clustersData.edges?.map(({ node }) => node)
  );

  const { account } = useParams();

  return (
    <>
      <SubHeader
        title="Cluster"
        actions={
          clusters.length !== 0 && (
            <Button
              variant="primary"
              content="Create Cluster"
              prefix={PlusFill}
              href={`/${account}/new-cluster`}
              LinkComponent={Link}
            />
          )
        }
      />
      {clusters.length > 0 && (
        <div className="pt-3xl flex flex-col gap-6xl">
          <div className="flex flex-col">
            <ClusterToolbar viewMode={viewMode} setViewMode={setViewMode} />
            <ClusterFilters
              appliedFilters={appliedFilters}
              setAppliedFilters={setAppliedFilters}
            />
          </div>
          <ResourceList mode={viewMode}>
            {clusters.map((cluster) => (
              <ResourceList.ResourceItem key={cluster.id}>
                <ResourceItem {...cluster} />
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
      {clusters.length === 0 && (
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
            heading="This is where you’ll manage your cluster "
            action={{
              content: 'Create new cluster',
              prefix: Plus,
              LinkComponent: Link,
              href: `/${account}/new-cluster`,
            }}
          >
            <p>You can create a new cluster and manage the listed cluster.</p>
          </EmptyState>
        </div>
      )}
    </>
  );
};

export const loader = async (ctx) => {
  const { data, errors } = await GQLServerHandler(ctx.request).listClusters({});

  if (errors) {
    logger.error(errors[0]);
  }

  return {
    clustersData: data || {},
  };
};

export default ClustersIndex;
