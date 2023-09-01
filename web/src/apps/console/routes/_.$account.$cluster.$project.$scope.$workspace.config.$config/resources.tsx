import { DotsThreeVerticalFill, Trash } from '@jengaicons/react';
import { useEffect, useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import { TextArea } from '~/components/atoms/input';
import OptionList from '~/components/atoms/option-list';
import { cn } from '~/components/utils';
import List from '~/console/components/list';

export interface IConfigItem {
  key: string;
  value: string;
}

interface IConfigItemExtended extends IConfigItem {
  delete: boolean;
  edit: boolean;
  insert: boolean;
  newvalue: string;
}

interface IResource {
  modifiedItems: { [key: string]: string };
  editItem: () => void;
  deleteItem: () => void;
  restoreItem: () => void;
}

const cc = (item: any) => ({
  '!text-text-critical line-through': item.delete,
  '!text-text-warning':
    !item.delete && item.newvalue && item.newvalue !== item.value,
  '!text-text-success': item.insert,
});

const ResourceItemExtraOptions = ({ onDelete = null, onRestore = null }) => {
  const [open, setOpen] = useState(false);
  useEffect(() => {
    console.log(open);
  }, [open]);
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      {/* @ts-ignore */}
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
      {/* @ts-ignore */}
      <OptionList.Content>
        {onRestore && (
          // @ts-ignore
          <OptionList.Item onSelect={onRestore}>
            <Trash size={16} />
            <span>Restore</span>
          </OptionList.Item>
        )}
        {onRestore && onDelete && <OptionList.Separator />}
        {onDelete && (
          // @ts-ignore
          <OptionList.Item
            className="!text-text-critical"
            onSelect={() => {
              // @ts-ignore
              onDelete();
              console.log('clicked');
            }}
          >
            <Trash size={16} />
            <span>Delete</span>
          </OptionList.Item>
        )}
      </OptionList.Content>
    </OptionList.Root>
  );
};

// @ts-ignore
const RenderItem = ({ item, onDelete, onEdit, onRestore, edit }) => {
  const [showDelete, setShowDelete] = useState(false);
  const [showRestore, setShowRestore] = useState(false);

  useEffect(() => {
    const timeout = setTimeout(() => {
      setShowRestore(
        item.value.delete ||
          (item.value.newvalue && item.value.newvalue !== item.value.value)
      );
      setShowDelete(!item.value.delete);
    }, 100);

    return () => {
      clearTimeout(timeout);
    };
  }, [item]);

  return (
    <div className="flex flex-col gap-xl">
      <div className="flex flex-row items-center gap-3xl">
        <div
          className={cn(
            'bodyMd-semibold text-text-default w-[300px]',
            cc(item.value)
          )}
        >
          {item.key}
        </div>
        <div className={cn('bodyMd text-text-soft flex-1', cc(item.value))}>
          {item.value.newvalue ? item.value.newvalue : item.value.value}
        </div>
        <ResourceItemExtraOptions
          onDelete={showDelete ? onDelete : null}
          onRestore={showRestore ? onRestore : null}
        />
      </div>
      {edit && !item.value.delete && (
        <TextArea
          label="value"
          resize={false}
          rows="4"
          value={item.value.newvalue ? item.value.newvalue : item.value.value}
          onClick={(e) => {
            e.stopPropagation();
          }}
          onChange={({ target }) => onEdit(target.value)}
        />
      )}
    </div>
  );
};

// @ts-ignore
const Resources = ({ modifiedItems, editItem, restoreItem, deleteItem }) => {
  const [selected, setSelected] = useState('');

  return (
    <List.Root>
      {Object.entries(modifiedItems).map(([key, value]) => {
        return (
          <List.Item
            key={key}
            pressed={selected === key}
            onClick={() => {
              setSelected((prev) => (prev === key ? '' : key));
            }}
            items={[
              {
                key: 1,
                className: 'flex-1',
                render: () => (
                  <RenderItem
                    edit={selected === key}
                    item={{ key, value }}
                    onDelete={() => deleteItem({ key, value })}
                    onEdit={(val: any) => editItem({ key, value }, val)}
                    onRestore={() => {
                      restoreItem({ key });
                      setSelected('');
                    }}
                  />
                ),
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

export default Resources;
