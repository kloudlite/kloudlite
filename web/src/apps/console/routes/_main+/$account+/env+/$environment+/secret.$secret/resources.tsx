import { Eye, PencilLine, SmileySad, Trash } from '~/console/components/icons';
import { useEffect, useState } from 'react';
import { cn } from '~/components/utils';
import AlertModal, { IAlertModal } from '~/console/components/alert-modal';
import ListGridView from '~/console/components/list-grid-view';
import NoResultsFound from '~/console/components/no-results-found';
import {
  ICSBase,
  ICSValueExtended,
  IModifiedItem,
  IShowDialog,
} from '~/console/components/types.d';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { ListTitle } from '~/console/components/console-list-components';
import ListV2 from '~/console/components/listV2';
import Handle from './handle';

// const RESOURCE_NAME = 'secret';

interface IResourceBase {
  modifiedItems: IModifiedItem;
  editItem: (item: ICSBase, value: string) => void;
  deleteItem: (item: ICSBase) => void;
  restoreItem: (item: ICSBase) => void;
  searchText: string;
}

interface IShowSecretDialog extends Omit<IAlertModal, 'setShow'> {
  data?: ICSBase;
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
  onShow: (item: ICSBase) => void;
};

const ValueComponent = ({
  item,
  onShow,
}: {
  item: ICSBase;
  onShow: (item: ICSBase) => void;
}) => {
  return (
    <div className={cn('bodyMd text-text-soft flex-1', cc(item.value))}>
      <div
        className="w-fit flex flew-row gap-xl items-center cursor-pointer"
        onClick={(e) => {
          e.stopPropagation();
          onShow(item);
        }}
      >
        <Eye size={16} /> •••••••••••••••
      </div>
    </div>
  );
};

// const GridView = ({
//   editItem,
//   restoreItem,
//   deleteItem,
//   onShow,
//   items,
// }: Omit<IResourceBase, 'modifiedItems' | 'searchText'> & {
//   onShow: (item: ICSBase) => void;
//   items: [string, ICSValueExtended][];
// }) => {
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
//                     onEdit={(val: string) => editItem({ key, value }, val)}
//                     onRestore={() => {
//                       restoreItem({ key, value });
//                       setSelected('');
//                     }}
//                     onShow={onShow}
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

const ListView = ({ items, onAction, onShow }: IResource) => {
  return (
    <ListV2.Root
      data={{
        headers: [
          {
            render: () => 'Key',
            name: 'key',
            className: 'w-[280px]',
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
                  <ListTitle
                    title={<span className={cc(item[1])}>{item[0]}</span>}
                  />
                ),
              },
              value: {
                render: () => (
                  <ValueComponent
                    item={{ key: item[0], value: item[1] }}
                    onShow={onShow}
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

const SecretItemResources = ({
  modifiedItems,
  searchText,
  deleteItem,
  editItem,
  restoreItem,
}: IResourceBase) => {
  const [showSecret, setShowSecret] = useState<IShowSecretDialog>({
    show: false,
    title: '',
    message: '',
  });

  const [items, setItems] = useState<[string, ICSValueExtended][]>([]);
  const [showHandleSecret, setShowHandleSecret] =
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

  const onShow = (item: ICSBase) => {
    setShowSecret({
      show: true,
      title: 'Confirmation',
      message: `Are you sure you want to view the value of '${item.key}'?`,
      footer: true,
      okText: 'Yes',
      cancelText: 'No',
      variant: 'primary',
      data: item,
    });
  };

  const props: IResource = {
    items,
    deleteItem,
    editItem,
    restoreItem,
    onShow,
    onAction: ({ action, item }) => {
      console.log('eee', action, item);

      const data = {
        key: item[0],
        value: item[1],
      };
      switch (action) {
        case 'edit':
          setShowHandleSecret({
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
      <AlertModal
        {...showSecret}
        setShow={setShowSecret}
        onSubmit={() => {
          setShowSecret({
            show: true,
            title: showSecret.data?.key,
            message: showSecret.data?.value.newvalue
              ? showSecret.data?.value.newvalue
              : showSecret.data?.value.value,
            footer: false,
            data: showSecret.data,
          });
        }}
      />
      <Handle
        show={showHandleSecret}
        setShow={setShowHandleSecret}
        onSubmit={(v, d) => {
          console.log(v, d);
          setShowHandleSecret(null);
          editItem({ key: v.key, value: d }, v.value);
        }}
        isUpdate
      />
    </>
  );
};

export default SecretItemResources;
