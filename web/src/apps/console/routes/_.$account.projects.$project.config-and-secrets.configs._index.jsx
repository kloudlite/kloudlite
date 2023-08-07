import {
  ArrowDown,
  ArrowUp,
  ArrowsDownUp,
  DiamondsFour,
  DotsThreeVerticalFill,
  PencilLine,
  Plus,
  Search,
  Trash,
} from '@jengaicons/react';
import { useEffect, useState } from 'react';
import * as Popup from '~/components/molecule/popup';
import * as AlertDialog from '~/components/molecule/alert-dialog';
import OptionList from '~/components/atoms/option-list';
import Toolbar from '~/components/atoms/toolbar';
import Pagination from '~/components/molecule/pagination';
import { cn } from '~/components/utils';
import { Button, IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { useOutletContext } from '@remix-run/react';
import ResourceList from '../components/resource-list';
import { dummyData } from '../dummy/data';

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
            Token name
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

const CRToolbar = () => {
  const [sortbyOptionListOpen, setSortybyOptionListOpen] = useState(false);
  return (
    <div>
      {/* Toolbar for md and up */}
      <div className="hidden md:flex">
        <Toolbar>
          <div className="w-full">
            <Toolbar.TextInput placeholder="Search" prefixIcon={Search} />
          </div>
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
        </Toolbar>
      </div>

      {/* Toolbar for mobile screen */}
      <div className="flex md:hidden">
        <Toolbar>
          <div className="flex-1">
            <Toolbar.TextInput placeholder="Search" prefixIcon={Search} />
          </div>
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
        </Toolbar>
      </div>
    </div>
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

// Project resouce item for grid and list mode
// mode param is passed from parent element
export const ResourceItem = ({ mode, item, onEdit, onDelete }) => {
  const [openExtra, setOpenExtra] = useState(false);
  const { name, entries, lastupdated } = item;

  const TitleComponent = () => (
    <>
      <div className="flex flex-row gap-md items-center">
        <div className="headingMd text-text-default">{name}</div>
      </div>
      <div className="bodyMd text-text-soft truncate">{lastupdated}</div>
    </>
  );

  const EntriesComponent = () => (
    <div className="bodyMd text-text-strong text-right w-[140px]">
      {entries}
    </div>
  );

  const OptionMenu = () => (
    <ResourceItemExtraOptions
      open={openExtra}
      setOpen={setOpenExtra}
      onEdit={() => {
        if (onEdit) onEdit(item);
      }}
      onDelete={() => {
        if (onDelete) onDelete(item);
      }}
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
            <div className="flex flex-col gap-sm w-[calc(100%-52px)] md:w-auto">
              {TitleComponent()}
            </div>
          </div>
          {OptionMenu()}
        </div>
        <div className="flex flex-col gap-md items-start">
          {EntriesComponent()}
        </div>
      </div>
    );
  };

  const listView = () => (
    <>
      <div className="hidden md:flex flex-row items-center justify-between gap-3xl w-full">
        <div className="flex flex-1 flex-row items-center gap-xl">
          <div className="flex flex-col gap-sm">{TitleComponent()}</div>
        </div>
        {EntriesComponent()}
        {OptionMenu()}
      </div>
      {gridView()}
    </>
  );

  if (mode === 'grid') return gridView();
  return listView();
};

const ProjectConfigIndex = () => {
  const [data, _setData] = useState(dummyData.projectConfig);
  const [currentPage, _setCurrentPage] = useState(1);
  const [itemsPerPage, _setItemsPerPage] = useState(15);
  const [totalItems, _setTotalItems] = useState(100);
  const [addConfig, setAddConfig] = useState(false);
  const [deleteConfig, setDeleteConfig] = useState(false);

  const [_subNavAction, setSubNavAction] = useOutletContext();

  useEffect(() => {
    setSubNavAction({
      action: () => {
        setAddConfig(true);
      },
    });
  }, []);

  return (
    <>
      <CRToolbar />
      <ResourceList mode="list">
        {data.map((d) => (
          <ResourceList.ResourceItem key={d.id} textValue={d.id}>
            <ResourceItem
              item={d}
              onDelete={(item) => {
                setDeleteConfig(item);
              }}
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
      {/* Popup dialog for adding config */}
      <Popup.PopupRoot show={addConfig} onOpenChange={setAddConfig}>
        <Popup.Header>
          <div className="flex flex-row gap-2xl items-center">
            <DiamondsFour size={20} /> <span>Add new config</span>
          </div>
        </Popup.Header>
        <form>
          <Popup.Content>
            <div className="flex flex-col gap-2xl">
              <TextInput label="Name" />
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
        show={deleteConfig}
        onOpenChange={setDeleteConfig}
      >
        <AlertDialog.Header>Delete config</AlertDialog.Header>
        <AlertDialog.Content>
          Are you sure you want to delete &quot;kloud-root-ca.crt&quot;.
        </AlertDialog.Content>
        <AlertDialog.Footer>
          <AlertDialog.Button variant="basic" content="Cancel" />
          <AlertDialog.Button
            variant="critical"
            content="Delete"
            onClick={(e) => {
              e.preventDefault();
              console.log(deleteConfig);
            }}
          />
        </AlertDialog.Footer>
      </AlertDialog.DialogRoot>
    </>
  );
};

export default ProjectConfigIndex;

export const handle = {
  subheaderAction: () => <Button content="Add new config" prefix={Plus} />,
};
