import { DotsThreeVerticalFill, Trash } from '@jengaicons/react';
import { useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import OptionList from '~/components/atoms/option-list';
import { cn } from '~/components/utils';

const ResourceItemExtraOptions = ({ open, setOpen, onDelete }) => {
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
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
const Resources = ({ mode, item, onDelete }) => {
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

export default Resources;
