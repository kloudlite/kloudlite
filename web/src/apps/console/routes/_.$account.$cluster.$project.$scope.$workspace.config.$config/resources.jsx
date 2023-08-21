import { DotsThreeVerticalFill, Trash } from '@jengaicons/react';
import { useEffect, useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import { TextArea } from '~/components/atoms/input';
import OptionList from '~/components/atoms/option-list';
import { cn } from '~/components/utils';
import List from '~/console/components/list';

const cc = (item) => ({
  '!text-text-critical line-through': item?.delete,
  '!text-text-warning':
    !item?.delete && item?.newvalue && item?.newvalue !== item.value,
  '!text-text-success': item?.insert,
});

const RenderItem = ({ item, onDelete, onEdit, onRestore }) => {
  const [showDelete, setShowDelete] = useState(false);
  const [showRestore, setShowRestore] = useState(false);

  useEffect(() => {
    let timeout = null;
    timeout = setTimeout(() => {
      setShowRestore(
        item?.delete || (item?.newvalue && item?.newvalue !== item.value)
      );
      setShowDelete(!item?.delete);
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
            cc(item)
          )}
        >
          {item.key}
        </div>
        <div className={cn('bodyMd text-text-soft flex-1', cc(item))}>
          {item?.newvalue ? item.newvalue : item.value}
        </div>
        <ResourceItemExtraOptions
          onDelete={showDelete ? onDelete : null}
          onRestore={showRestore ? onRestore : null}
        />
      </div>
      {item?.edit && !item?.delete && (
        <TextArea
          label="value"
          resize={false}
          rows="4"
          value={item?.newvalue ? item.newvalue : item.value}
          onClick={(e) => {
            e.stopPropagation();
          }}
          onChange={({ target }) => onEdit(target.value)}
        />
      )}
    </div>
  );
};

const Resources = ({
  originalItems = [],
  modifiedItems = [],
  setModifiedData,
}) => {
  const deleteItem = (item) => {
    if (originalItems.find((or) => or.key === item.key)) {
      setModifiedData(
        modifiedItems.map((mi) => {
          if (mi.key === item.key) {
            return { ...mi, delete: true, insert: false };
          }
          return mi;
        })
      );
    } else {
      setModifiedData(modifiedItems.filter((mi) => mi.key !== item.key));
    }
  };

  const editItem = (item, value) => {
    setModifiedData(
      modifiedItems.map((mi) => {
        if (mi.key === item.key) {
          return { ...mi, newvalue: value };
        }
        return mi;
      })
    );
  };

  const restoreItem = (item) => {
    const orgItem = originalItems.find((or) => or.key === item.key);
    if (orgItem) {
      setModifiedData(
        modifiedItems.map((mi) => {
          if (mi.key === item.key) {
            return orgItem;
          }
          return mi;
        })
      );
    }
  };

  return (
    <List.Root>
      {modifiedItems.map((item) => (
        <List.Item
          key={item.key}
          pressed={item?.edit}
          onClick={() => {
            setModifiedData([
              ...modifiedItems.map((mi) => {
                if (mi.key === item.key) return { ...mi, edit: !mi?.edit };
                return { ...mi, edit: false };
              }),
            ]);
          }}
          items={[
            {
              key: 1,
              className: 'flex-1',
              render: () => (
                <RenderItem
                  item={item}
                  onDelete={() => deleteItem(item)}
                  onEdit={(value) => editItem(item, value)}
                  onRestore={() => restoreItem(item)}
                />
              ),
            },
          ]}
        />
      ))}
    </List.Root>
  );
};

const ResourceItemExtraOptions = ({ onDelete = null, onRestore = null }) => {
  const [open, setOpen] = useState(false);
  useEffect(() => {
    console.log(open);
  }, [open]);
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
        {onRestore && (
          <OptionList.Item onSelect={onRestore}>
            <Trash size={16} />
            <span>Restore</span>
          </OptionList.Item>
        )}
        {onRestore && onDelete && <OptionList.Separator />}
        {onDelete && (
          <OptionList.Item
            className="!text-text-critical"
            onSelect={() => {
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

export default Resources;
