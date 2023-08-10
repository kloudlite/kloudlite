import { DotsThreeVerticalFill, Trash } from '@jengaicons/react';
import { useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import { TextArea } from '~/components/atoms/input';
import OptionList from '~/components/atoms/option-list';
import { cn } from '~/components/utils';

const Resources = ({ item, onDelete }) => {
  const { key, value } = item;
  const [edit, setEdit] = useState(false);

  const [openExtra, setOpenExtra] = useState(false);

  const StartComponent = () => (
    <div className="headingMd text-text-default">{key}</div>
  );

  const EndComponent = () => (
    <div className="bodyMd text-text-strong">{value}</div>
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

  const listView = () => (
    <div
      className="hidden md:flex flex-col gap-xl md:w-full"
      onClick={() => setEdit(true)}
    >
      <div className="flex flex-row items-center gap-3xl">
        <div
          className={cn('flex flex-col gap-sm w-[300px]', {
            'flex-1': edit,
          })}
        >
          {StartComponent()}
        </div>
        {!edit && <div className="flex flex-col flex-1">{EndComponent()}</div>}
        {OptionMenu()}
      </div>
      {edit && (
        <div className="flex flex-col gap-md">
          <TextArea
            label="Value"
            rows="4"
            resize={false}
            onClick={(e) => {
              e.stopPropagation();
              e.target.focus();
            }}
            onMouseDown={(e) => {
              e.stopPropagation();
            }}
            onPointerDown={(e) => {
              e.stopPropagation();
            }}
          />
        </div>
      )}
    </div>
  );

  return listView();
};

const ResourceItemExtraOptions = ({ open, setOpen, onDelete }) => {
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
        <OptionList.Item className="!text-text-critical" onSelect={onDelete}>
          <Trash size={16} />
          <span>Delete</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList>
  );
};

export default Resources;
