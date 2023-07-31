import { useEffect, useState } from 'react';
import { Link } from '@remix-run/react';
import {
  ArrowDown,
  ArrowUp,
  ArrowsDownUp,
  CaretDownFill,
  Cloud,
  Copy,
  CopySimple,
  DotsThreeVerticalFill,
  Info,
  List,
  PencilLine,
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
import Pagination from '~/components/molecule/pagination';
import { AnimatePresence, motion } from 'framer-motion';
import * as Chips from '~/components/atoms/chips';
import { cn } from '~/components/utils';
import { ChipGroupPaddingTop } from '~/design-system/tailwind-base';
import { Badge } from '~/components/atoms/badge';
import * as Popup from '~/components/molecule/popup';
import * as AlertDialog from '~/components/molecule/alert-dialog';
import { TextInput } from '~/components/atoms/input';
import * as SelectInput from '~/components/atoms/select';
import ResourceList from '../components/resource-list';
import { EmptyState } from '../components/empty-state';
import ScrollArea from '../components/scroll-area';

const CloudProviderList = [
  {
    name: 'Lobster early',
    id: 'lobster-early-kloudlite-app1',
    providerRegion: 'Amazon Web Services',
    author: 'Reyan updated the project',
    status: 'Verified',
    lastupdated: '3 days ago',
  },
  {
    name: 'Lobster early',
    id: 'lobster-early-kloudlite-app2',
    providerRegion: 'Amazon Web Services',
    author: 'Reyan updated the project',
    status: 'Verified',
    lastupdated: '3 days ago',
  },
  {
    name: 'Lobster early',
    id: 'lobster-early-kloudlite-app3',
    providerRegion: 'Amazon Web Services',
    author: 'Reyan updated the project',
    status: 'Verified',
    lastupdated: '3 days ago',
  },
  {
    name: 'Lobster early',
    id: 'lobster-early-kloudlite-app4',
    providerRegion: 'Amazon Web Services',
    author: 'Reyan updated the project',
    status: 'Verified',
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
  status,
  lastupdated,
  author,
  onEdit,
  onDelete,
}) => {
  const [openExtra, setOpenExtra] = useState(false);

  const ThumbnailComponent = () => (
    <span className="self-start">
      <Cloud size={20} />
    </span>
  );

  const TitleComponent = () => (
    <>
      <div className="flex flex-row gap-md items-center">
        <div className="headingMd text-text-default">{name}</div>
        {/* <div className="w-lg h-lg bg-icon-primary rounded-full" /> */}
      </div>
      <div className="bodyMd text-text-soft truncate">{id}</div>
    </>
  );

  const ClusterComponent = () => (
    <>
      <div className="w-[120px]">
        <Badge label={status} icon={Info} />
      </div>
      <div className="bodyMd text-text-strong w-[200px] flex flex-row items-center gap-lg">
        <svg
          width="16"
          height="12"
          viewBox="0 0 16 12"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M5.6687 1.20159L5.28551 0.269004C5.21808 0.104897 5.03445 0.0217588 4.86664 0.0793625L2.53864 0.878497C2.4347 0.914176 2.35486 0.998537 2.32494 1.10428L0.0127094 9.27688C-0.0224498 9.40115 0.0164463 9.5348 0.115669 9.61747C0.976552 10.3347 2.80426 11.5903 4.52704 11.9411C4.66055 11.9683 4.79463 11.905 4.86513 11.7884L5.35357 10.9805C5.39679 10.9071 5.35383 10.813 5.27021 10.7964C4.80887 10.7051 4.35216 10.5889 3.90212 10.448C3.54878 10.3374 3.35199 9.96132 3.46258 9.60798C3.57317 9.25464 3.94926 9.05786 4.3026 9.16845C5.46468 9.53216 6.67623 9.71299 7.89386 9.70446H7.90326C9.12088 9.71299 10.3324 9.53216 11.4945 9.16846C11.8478 9.05787 12.2239 9.25465 12.3345 9.60799C12.4451 9.96134 12.2483 10.3374 11.895 10.448C11.5002 10.5716 11.1002 10.6762 10.6966 10.7617C10.5597 10.7907 10.4907 10.9462 10.5636 11.0656L10.6006 11.1262L10.993 11.7649C11.0716 11.8928 11.2254 11.9533 11.3688 11.9091C12.5367 11.5495 14.7073 10.7043 15.8986 9.61923C15.9904 9.53566 16.0216 9.40608 15.9851 9.28744L13.4691 1.10042C13.4373 0.996902 13.3574 0.915134 13.2547 0.880889L10.822 0.0699893C10.6621 0.0166939 10.4875 0.0904079 10.4141 0.242162L10.132 0.826162L10.054 0.998844C9.99898 1.12053 10.0764 1.26089 10.2083 1.28179C10.7726 1.37122 11.3309 1.49475 11.8793 1.65191C12.2352 1.7539 12.4411 2.12511 12.3391 2.48103C12.2371 2.83695 11.8659 3.0428 11.51 2.94081C10.3435 2.60654 9.12645 2.44013 7.90276 2.44798H7.89416C6.67046 2.44014 5.4534 2.60655 4.28688 2.94083C3.93096 3.04282 3.55975 2.83697 3.45776 2.48105C3.35577 2.12513 3.56162 1.75393 3.91753 1.65193C4.47803 1.49132 5.04874 1.36583 5.62572 1.27596C5.6605 1.27054 5.68145 1.23439 5.6687 1.20159ZM6.48193 7.09922C6.48193 7.56863 6.1014 7.94916 5.63199 7.94916C5.16259 7.94916 4.78206 7.56863 4.78206 7.09922C4.78206 6.62982 5.16259 6.24929 5.63199 6.24929C6.1014 6.24929 6.48193 6.62982 6.48193 7.09922ZM11.0149 7.09922C11.0149 7.56863 10.6343 7.94916 10.1649 7.94916C9.69554 7.94916 9.31501 7.56863 9.31501 7.09922C9.31501 6.62982 9.69554 6.24929 10.1649 6.24929C10.6343 6.24929 11.0149 6.62982 11.0149 7.09922Z"
            fill="#4B5563"
          />
        </svg>
        {providerRegion}
      </div>
    </>
  );

  const AuthorComponent = () => (
    <>
      <div className="bodyMd text-text-strong">{author}</div>
      <div className="bodyMd text-text-soft">{lastupdated}</div>
    </>
  );

  const OptionMenu = () => (
    <ResourceItemExtraOptions
      open={openExtra}
      setOpen={setOpenExtra}
      onEdit={onEdit}
      onDelete={onDelete}
    />
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
        <div className="flex flex-col w-[200px]">{AuthorComponent()}</div>
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
          <OptionList.Item
            key={provider.id}
            onSelect={(e) => e.preventDefault()}
          >
            <div className="flex flex-row gap-xl">
              <CopySimple size={16} />
              {provider.content}
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

const ResourceItemExtraOptions = ({ open, setOpen, onEdit, onDelete }) => {
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
        <OptionList.Item onSelect={onEdit}>
          <PencilLine size={16} />
          <span>Edit</span>
        </OptionList.Item>
        <OptionList.Separator />
        <OptionList.Item className="!text-text-critical" onSelect={onDelete}>
          <Trash size={16} />
          <span>Delete</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList>
  );
};

const CloudProvidersIndex = () => {
  const [cloudProviders, _setCloudProviders] = useState(CloudProviderList);
  const [appliedFilters, setAppliedFilters] = useState(AppliedFilters);
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(15);
  const [totalItems, setTotalItems] = useState(100);
  const [viewMode, setViewMode] = useState('list');
  const [addProviderPopup, setAddProviderPopup] = useState(false);
  const [editProviderPopup, setEditProviderPopup] = useState(false);
  const [deleteProviderPopup, setDeleteProviderPopup] = useState(false);

  return (
    <>
      <SubHeader
        title="Cloud Providers"
        actions={
          cloudProviders.length !== 0 && (
            <Button
              variant="primary"
              content="Create Cloud Provider"
              prefix={PlusFill}
              type="button"
              onClick={() => setAddProviderPopup(true)}
            />
          )
        }
      />
      {cloudProviders.length > 0 && (
        <div className="pt-3xl flex flex-col gap-6xl">
          <div className="flex flex-col">
            <ClusterToolbar viewMode={viewMode} setViewMode={setViewMode} />
            <ClusterFilters
              appliedFilters={appliedFilters}
              setAppliedFilters={setAppliedFilters}
            />
          </div>
          <ResourceList mode={viewMode}>
            {cloudProviders.map((cluster) => (
              <ResourceList.ResourceItem key={cluster.id}>
                <ResourceItem
                  {...cluster}
                  onEdit={() => setEditProviderPopup(true)}
                  onDelete={() => setDeleteProviderPopup(true)}
                />
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
      {cloudProviders.length === 0 && (
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
            heading="This is the place where you will oversees the Cloud Provider."
            action={{
              content: 'Create new cloud provider',
              prefix: Plus,
              LinkComponent: Link,
              href: '/new-project',
            }}
          >
            <p>
              You have the option to include a new Cloud Provider and oversee
              the existing Cloud Provider.
            </p>
          </EmptyState>
        </div>
      )}

      {/* Popup dialog for adding cloud provider */}
      <Popup.PopupRoot
        show={addProviderPopup}
        onOpenChange={setAddProviderPopup}
      >
        <Popup.Header>Add new cloud provider</Popup.Header>
        <form>
          <Popup.Content>
            <div className="flex flex-col gap-2xl">
              <TextInput label="Name" />
              <TextInput
                label="Handle"
                suffixIcon={Info}
                extra={
                  <Button
                    size="md"
                    variant="primary-plain"
                    content="Edit"
                    href="#"
                    LinkComponent={Link}
                  />
                }
              />
              <SelectInput.Select value="" label="Provider">
                <SelectInput.Option>--Select--</SelectInput.Option>
              </SelectInput.Select>
              <TextInput label="Access Key ID" />
              <TextInput label="Secret Access Key" />
            </div>
          </Popup.Content>
          <Popup.Footer>
            <Popup.Button content="Cancel" variant="basic" />
            <Popup.Button content="Add" variant="primary" />
          </Popup.Footer>
        </form>
      </Popup.PopupRoot>

      {/* Popup dialog for editing cloud provider */}
      <Popup.PopupRoot
        show={editProviderPopup}
        onOpenChange={setEditProviderPopup}
      >
        <Popup.Header>Add new cloud provider</Popup.Header>
        <form>
          <Popup.Content>
            <div className="flex flex-col gap-2xl">
              <TextInput label="Name" />
              <TextInput label="Access Key ID" />
              <TextInput label="Secret Access Key" />
            </div>
          </Popup.Content>
          <Popup.Footer>
            <Popup.Button content="Cancel" variant="basic" />
            <Popup.Button content="Add" variant="primary" />
          </Popup.Footer>
        </form>
      </Popup.PopupRoot>

      {/* Alert Dialog for deleting cloud provider */}
      <AlertDialog.DialogRoot
        show={deleteProviderPopup}
        onOpenChange={setDeleteProviderPopup}
      >
        <AlertDialog.Header>Delete Cloud Provider</AlertDialog.Header>
        <AlertDialog.Content>
          Are you sure you want to delete 'kloud-root-ca.crt".
        </AlertDialog.Content>
        <AlertDialog.Footer>
          <AlertDialog.Button variant="basic" content="Cancel" />
          <AlertDialog.Button variant="critical" content="Delete" />
        </AlertDialog.Footer>
      </AlertDialog.DialogRoot>
    </>
  );
};

export default CloudProvidersIndex;
