import { DotsThreeVerticalFill, SmileySad, Trash } from '@jengaicons/react';
import { useEffect, useState } from 'react';
import AnimateHide from '~/components/atoms/animate-hide';
import { IconButton } from '~/components/atoms/button';
import { TextArea } from '~/components/atoms/input';
import OptionList from '~/components/atoms/option-list';
import { cn, generateKey } from '~/components/utils';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import NoResultsFound from '~/console/components/no-results-found';
import {
  ICSBase,
  ICSValueExtended,
  IModifiedItem,
} from '~/console/components/types.d';

const RESOURCE_NAME = 'config';

interface IRenderItem {
  item: ICSBase;
  onDelete: () => void;
  onEdit: (value: string) => void;
  onRestore: () => void;
  edit: boolean;
  listMode?: boolean;
}

interface IResource {
  modifiedItems: IModifiedItem;
  editItem: (item: ICSBase, value: string) => void;
  deleteItem: (item: ICSBase) => void;
  restoreItem: (item: ICSBase) => void;
  searchText: string;
}

interface IResourceItemExtraOptions {
  onDelete: (() => void) | null;
  onRestore: (() => void) | null;
}

const cc = (item: ICSValueExtended): string =>
  cn({
    '!text-text-critical line-through': item.delete,
    '!text-text-warning':
      !item.delete && item.newvalue != null && item.newvalue !== item.value,
    '!text-text-success': item.insert,
  });

const ResourceItemExtraOptions = ({
  onDelete,
  onRestore,
}: IResourceItemExtraOptions) => {
  const [open, setOpen] = useState(false);

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
        {onRestore && (
          <OptionList.Item onClick={onRestore}>
            <Trash size={16} />
            <span>Restore</span>
          </OptionList.Item>
        )}
        {onRestore && onDelete && <OptionList.Separator />}
        {onDelete && (
          <OptionList.Item
            className="!text-text-critical"
            onClick={() => {
              onDelete();
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

const ValueComponent = ({
  item,
  listMode,
}: {
  item: ICSBase;
  listMode: boolean;
}) => {
  return (
    <div
      className={cn(
        'bodyMd text-text-soft flex-1',
        {
          'line-clamp-2': !listMode,
          'line-clamp-1': listMode,
        },
        cc(item.value)
      )}
    >
      {item.value.newvalue != null ? item.value.newvalue : item.value.value}
    </div>
  );
};
const RenderItem = ({
  item,
  onDelete,
  onEdit,
  onRestore,
  edit,
  listMode = true,
}: IRenderItem) => {
  const [showDelete, setShowDelete] = useState(false);
  const [showRestore, setShowRestore] = useState(false);

  useEffect(() => {
    const timeout = setTimeout(() => {
      setShowRestore(
        item.value.delete ||
          (item.value.newvalue != null &&
            item.value.newvalue !== item.value.value)
      );
      setShowDelete(!item.value.delete);
    }, 100);

    return () => {
      clearTimeout(timeout);
    };
  }, [item]);

  return (
    <div className={cn('flex flex-col')}>
      <div className="flex flex-col gap-lg">
        <div className={cn('flex flex-row items-center gap-3xl')}>
          <div
            className={cn(
              'bodyMd-semibold text-text-default w-[300px]',
              cc(item.value)
            )}
          >
            {item.key}
          </div>
          {listMode && <ValueComponent item={item} listMode={listMode} />}
          <ResourceItemExtraOptions
            onDelete={showDelete ? onDelete : null}
            onRestore={showRestore ? onRestore : null}
          />
        </div>
        {!listMode && <ValueComponent item={item} listMode={listMode} />}
      </div>
      <AnimateHide show={edit && !item.value.delete}>
        <div
          className={cn({
            'pt-xl': listMode,
            'pt-3xl': !listMode,
          })}
        >
          <TextArea
            label="value"
            resize={false}
            rows="4"
            value={
              item.value.newvalue != null
                ? item.value.newvalue
                : item.value.value
            }
            onClick={(e) => {
              e.stopPropagation();
            }}
            onKeyDown={(e) => {
              e.stopPropagation();
            }}
            onChange={({ target }) => onEdit(target.value)}
          />
        </div>
      </AnimateHide>
    </div>
  );
};

const GridView = ({
  editItem,
  restoreItem,
  deleteItem,
  items,
}: Omit<IResource, 'modifiedItems' | 'searchText'> & {
  items: [string, ICSValueExtended][];
}) => {
  const [selected, setSelected] = useState('');

  return (
    <Grid.Root>
      {items.map(([key, value], index) => {
        const keyPrefix = `${RESOURCE_NAME}-${key}-${index}`;
        return (
          <Grid.Column
            key={key}
            pressed={selected === key}
            onClick={() => {
              setSelected((prev) => (prev === key ? '' : key));
            }}
            className="h-fit min-h-[118px]"
            rows={[
              {
                key: generateKey(keyPrefix + key),
                render: () => (
                  <RenderItem
                    edit={selected === key}
                    item={{ key, value }}
                    onDelete={() => deleteItem({ key, value })}
                    onEdit={(val: string) => editItem({ key, value }, val)}
                    onRestore={() => {
                      restoreItem({ key, value });
                      setSelected('');
                    }}
                    listMode={false}
                  />
                ),
              },
            ]}
          />
        );
      })}
    </Grid.Root>
  );
};

const ListView = ({
  editItem,
  restoreItem,
  deleteItem,
  items,
}: Omit<IResource, 'modifiedItems' | 'searchText'> & {
  items: [string, ICSValueExtended][];
}) => {
  const [selected, setSelected] = useState('');

  return (
    <List.Root>
      {items.map(([key, value]) => {
        return (
          <List.Row
            key={key}
            pressed={selected === key}
            onClick={() => {
              setSelected((prev) => (prev === key ? '' : key));
            }}
            columns={[
              {
                key: 1,
                className: 'flex-1',
                render: () => (
                  <RenderItem
                    edit={selected === key}
                    item={{ key, value }}
                    onDelete={() => deleteItem({ key, value })}
                    onEdit={(val: string) => editItem({ key, value }, val)}
                    onRestore={() => {
                      restoreItem({ key, value });
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
const ConfigItemResources = ({
  modifiedItems,
  searchText,
  deleteItem,
  editItem,
  restoreItem,
}: IResource) => {
  const [items, setItems] = useState<[string, ICSValueExtended][]>([]);

  useEffect(() => {
    setItems(
      Object.entries(modifiedItems).filter(([key, _value]) => {
        if (
          key.toLowerCase().includes(searchText.toLowerCase()) ||
          !searchText
        ) {
          return true;
        }
        return false;
      })
    );
  }, [searchText, modifiedItems]);

  const props = {
    items,
    deleteItem,
    editItem,
    restoreItem,
  };
  return (
    <>
      {(!searchText || (searchText && items.length > 0)) && (
        <ListGridView
          listView={<ListView {...props} />}
          gridView={<GridView {...props} />}
        />
      )}
      {!!searchText && items.length === 0 && (
        <NoResultsFound
          title="No results found"
          subtitle="Try changing the filters or search terms for this view."
          image={<SmileySad size={40} />}
        />
      )}
    </>
  );
};

export default ConfigItemResources;
