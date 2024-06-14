import { PencilLine, SmileySad, Trash } from '~/console/components/icons';
import { useEffect, useState } from 'react';
import { cn } from '~/components/utils';
import ListGridView from '~/console/components/list-grid-view';
import NoResultsFound from '~/console/components/no-results-found';
import {
  ICSBase,
  ICSValueExtended,
  IModifiedItem,
  IShowDialog,
} from '~/console/components/types.d';
import ListV2 from '~/console/components/listV2';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { ListItem } from '~/console/components/console-list-components';
import Handle from './handle';

interface IResourceBase {
  modifiedItems: IModifiedItem;
  editItem: (item: ICSBase, value: string) => void;
  deleteItem: (item: ICSBase) => void;
  restoreItem: (item: ICSBase) => void;
  searchText: string;
}

const cc = (item: ICSValueExtended): string =>
  cn({
    '!text-text-critical line-through': item.delete,
    '!text-text-warning':
      !item.delete && item.newvalue != null && item.newvalue !== item.value,
    '!text-text-success': item.insert,
  });

type OnAction = ({
  action,
  item,
}: {
  action: 'delete' | 'edit' | 'restore';
  item: [string, ICSValueExtended];
}) => void;

type IExtraButton = {
  onAction: OnAction;
  item: [string, ICSValueExtended];
};

const ExtraButton = ({ onAction, item }: IExtraButton) => {
  const iconSize = 16;

  return item[1].newvalue || item[1].delete ? (
    <ResourceExtraAction
      options={[
        {
          label: 'Restore',
          type: 'item',
          key: 'restore',
          icon: <PencilLine size={iconSize} />,
          onClick: () => onAction({ action: 'restore', item }),
        },
        {
          label: 'Edit',
          type: 'item',
          key: 'edit',
          icon: <PencilLine size={iconSize} />,
          onClick: () => onAction({ action: 'edit', item }),
        },
        {
          label: 'Delete',
          type: 'item',
          key: 'delete',
          icon: <Trash size={iconSize} />,
          onClick: () => onAction({ action: 'delete', item }),
          className: '!text-text-critical',
        },
      ]}
    />
  ) : (
    <ResourceExtraAction
      options={[
        {
          label: 'Edit',
          type: 'item',
          key: 'edit',
          icon: <PencilLine size={iconSize} />,
          onClick: () => onAction({ action: 'edit', item }),
        },
        {
          label: 'Delete',
          type: 'item',
          key: 'delete',
          icon: <Trash size={iconSize} />,
          onClick: () => onAction({ action: 'delete', item }),
          className: '!text-text-critical',
        },
      ]}
    />
  );
};

type IResource = Omit<IResourceBase, 'modifiedItems' | 'searchText'> & {
  items: [string, ICSValueExtended][];
  onAction: OnAction;
};

// const GridView = ({ editItem, restoreItem, deleteItem, items }: IResource) => {
//   const [selected, setSelected] = useState('');

//   return (
//     <Grid.Root>
//       {items.map(([key, value], index) => {
//         const keyPrefix = `${RESOURCE_NAME}-${key}-${index}`;
//         return (
//           <Grid.Column
//             key={key}
//             pressed={selected === key}
//             onClick={() => {
//               setSelected((prev) => (prev === key ? '' : key));
//             }}
//             className="h-fit min-h-[118px]"
//             rows={[
//               {
//                 key: generateKey(keyPrefix + key),
//                 render: () => (
//                   <RenderItem
//                     edit={selected === key}
//                     item={{ key, value }}
//                     onDelete={() => deleteItem({ key, value })}
//                     // onEdit={(val: string) => editItem({ key, value }, val)}
//                     onRestore={() => {
//                       restoreItem({ key, value });
//                       setSelected('');
//                     }}
//                     listMode={false}
//                   />
//                 ),
//               },
//             ]}
//           />
//         );
//       })}
//     </Grid.Root>
//   );
// };

const ListView = ({ items, onAction }: IResource) => {
  return (
    <ListV2.Root
      data={{
        headers: [
          {
            render: () => 'Key',
            name: 'key',
            className: 'w-[180px]',
          },
          {
            render: () => 'Value',
            name: 'value',
            className: 'flex-1',
          },
          {
            render: () => '',
            name: 'action',
            className: 'w-[24px]',
          },
        ],
        rows: items.map((item) => {
          return {
            columns: {
              key: {
                render: () => (
                  <ListItem
                    data={<span className={cc(item[1])}>{item[0]}</span>}
                  />
                ),
              },
              value: {
                render: () => (
                  <ListItem
                    data={
                      <span className={cc(item[1])}>
                        {item[1].newvalue || item[1].value}
                      </span>
                    }
                  />
                ),
              },
              action: {
                render: () => <ExtraButton onAction={onAction} item={item} />,
              },
            },
          };
        }),
      }}
    />
  );
};

const ConfigItemResources = ({
  modifiedItems,
  searchText,
  deleteItem,
  editItem,
  restoreItem,
}: IResourceBase) => {
  const [items, setItems] = useState<[string, ICSValueExtended][]>([]);
  const [showHandleConfig, setShowHandleConfig] =
    useState<IShowDialog<IModifiedItem>>(null);

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

  const props: IResource = {
    items,
    deleteItem,
    editItem,
    restoreItem,
    onAction: ({ action, item }) => {
      console.log('eee', action, item);

      const data = {
        key: item[0],
        value: item[1],
      };
      switch (action) {
        case 'edit':
          setShowHandleConfig({
            type: 'edit',
            data: { [item[0]]: item[1] },
          });
          break;
        case 'delete':
          deleteItem(data);
          break;
        case 'restore':
          restoreItem(data);
          break;
        default:
      }
    },
  };
  return (
    <>
      {(!searchText || (searchText && items.length > 0)) && (
        <ListGridView
          listView={<ListView {...props} />}
          gridView={<ListView {...props} />}
        />
      )}
      {!!searchText && items.length === 0 && (
        <NoResultsFound
          title="No results found"
          subtitle="Try changing the filters or search terms for this view."
          image={<SmileySad size={40} />}
        />
      )}
      <Handle
        show={showHandleConfig}
        setShow={setShowHandleConfig}
        onSubmit={(v, d) => {
          console.log(v, d);
          setShowHandleConfig(null);
          editItem({ key: v.key, value: d }, v.value);
        }}
        isUpdate
      />
    </>
  );
};

export default ConfigItemResources;
