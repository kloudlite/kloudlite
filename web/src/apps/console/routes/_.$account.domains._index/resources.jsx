import { DotsThreeVerticalFill, Folder, Info, Trash } from '@jengaicons/react';
import { useState } from 'react';
import { Badge } from '~/components/atoms/badge';
import { Button, IconButton } from '~/components/atoms/button';
import OptionList from '~/components/atoms/option-list';
import { cn } from '~/components/utils';

const ResourceItemExtraOptions = ({ open, setOpen, onDelete }) => {
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <IconButton
          variant="plain"
          icon={<DotsThreeVerticalFill />}
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
        <OptionList.Item className="!text-text-critical" onSelect={onDelete}>
          <Trash size={16} />
          <span>Delete</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList.Root>
  );
};

// Project resouce item for grid and list mode
// mode param is passed from parent element
const Resources = (
  { mode = '', item, onDelete = (_) => _, onEdit = (_) => _ } = { item: {} }
) => {
  const { name, status, cluster, activity, lastupdated } = item;

  const [openExtra, setOpenExtra] = useState(false);
  const StartComponent = () => (
    <div className="flex flex-row gap-md items-center">
      <div className="headingMd text-text-default">{name}</div>
    </div>
  );

  const MiddleComponent = () => (
    <>
      <div className="bodyMd text-text-strong text-start w-[200px]">
        {cluster}
      </div>
      <div className="bodyMd text-text-strong text-start w-[120px]">
        <Badge
          label={status}
          icon={<Info />}
          type={status === 'Verified' ? 'neutral' : 'critical'}
        />
      </div>
    </>
  );

  const EndComponent = () => (
    <>
      <div className="bodyMd text-text-strong">{activity}</div>
      <div className="bodyMd text-text-soft">{lastupdated}</div>
    </>
  );

  const OptionMenu = () => (
    <ResourceItemExtraOptions
      open={openExtra}
      setOpen={setOpenExtra}
      onDelete={() => onDelete(item)}
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
              {StartComponent()}
            </div>
          </div>
          {OptionMenu()}
        </div>
        <div className="flex flex-col gap-md items-start">
          {MiddleComponent()}
        </div>
        <div className="flex flex-col items-start">{EndComponent()}</div>
      </div>
    );
  };

  const listView = () => (
    <>
      <div className="hidden md:flex flex-row items-center justify-between gap-3xl w-full">
        <div className="flex flex-1 flex-row items-center gap-xl">
          <div className="flex flex-col gap-sm">{StartComponent()}</div>
        </div>
        {MiddleComponent()}
        <div className="flex flex-col w-[200px]">{EndComponent()}</div>
        {OptionMenu()}
      </div>
      {gridView()}
    </>
  );

  if (mode === 'grid') return gridView();
  return listView();
};

export const ResourceProjectItem = ({ item, onSelect }) => {
  const { name, selected } = item;

  return (
    <div
      className="h-[44px] py-lg px-2xl flex flex-row gap-lg items-center group"
      onClick={onSelect}
    >
      <Folder size={15} />
      <div className="headingSm text-text-default flex-1">{name}</div>
      <Button
        variant="plain"
        size="sm"
        content={selected ? 'Selected' : 'Select'}
        onClick={(e) => {
          e.stopPropagation();
          if (onSelect) onSelect(e);
        }}
        disabled={selected}
        className={cn({
          'hidden group-hover:flex': !selected,
          flex: selected,
        })}
        onMouseDown={(e) => {
          e.stopPropagation();
        }}
        onPointerDown={(e) => {
          e.stopPropagation();
        }}
      />
    </div>
  );
};

export default Resources;
