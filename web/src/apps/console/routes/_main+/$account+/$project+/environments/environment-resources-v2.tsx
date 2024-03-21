import { Copy, GearSix } from '@jengaicons/react';
import { Link, useOutletContext, useParams } from '@remix-run/react';
import { useState } from 'react';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListItem,
  ListTitle,
  listClass,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { IEnvironments } from '~/console/server/gql/queries/environment-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { SyncStatusV2 } from '~/console/components/sync-status';
import { IAccountContext } from '~/console/routes/_main+/$account+/_layout';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import { IProjectContext } from '~/console/routes/_main+/$account+/$project+/_layout';
import ListV2 from '~/console/components/listV2';
import CloneEnvironment from './clone-environment';

const RESOURCE_NAME = 'environment';
type BaseType = ExtractNodeType<IEnvironments>;

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: parseName(item),
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

type OnAction = ({ action, item }: { action: 'clone'; item: BaseType }) => void;

type IExtraButton = {
  onAction: OnAction;
  item: BaseType;
};

const ExtraButton = ({ item, onAction }: IExtraButton) => {
  const { account, project } = useParams();
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Clone',
          icon: <Copy size={16} />,
          type: 'item',
          key: 'clone',
          onClick: () => onAction({ action: 'clone', item }),
        },
        {
          label: 'Settings',
          icon: <GearSix size={16} />,
          type: 'item',
          to: `/${account}/${project}/env/${parseName(item)}/settings/general`,
          key: 'settings',
        },
      ]}
    />
  );
};

interface IResource {
  items: BaseType[];
  onAction: OnAction;
}

const GridView = ({ items = [], onAction }: IResource) => {
  const { account, project } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/${project}/env/${id}`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                className: listClass.title,
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={<ExtraButton item={item} onAction={onAction} />}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: listClass.author,
                render: () => (
                  <ListItem
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
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

const ListView = ({ items, onAction }: IResource) => {
  const { account, project } = useParams();

  return (
    <ListV2.Root
      linkComponent={Link}
      data={{
        headers: [
          {
            render: () => (
              <div className="flex flex-row">
                <span className="w-[48px]" />
                Name
              </div>
            ),
            name: 'name',
            className: 'w-[180px]',
          },
          {
            render: () => 'Status',
            name: 'status',
            className: 'flex-1 min-w-[30px] flex items-center justify-center',
          },
          {
            render: () => 'Environmet',
            name: 'environment',
            className: 'w-[180px]',
          },
          {
            render: () => 'Updated',
            name: 'updated',
            className: 'w-[180px]',
          },
          {
            render: () => '',
            name: 'action',
            className: 'w-[24px]',
          },
        ],
        rows: items.map((i) => {
          const { name, id, updateInfo } = parseItem(i);
          return {
            columns: {
              name: {
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              status: {
                render: () => <SyncStatusV2 item={i} />,
              },
              environment: {
                render: () => <ListItem data={i.spec?.routing?.mode} />,
              },
              updated: {
                render: () => (
                  <ListItem
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              action: {
                render: () => <ExtraButton item={i} onAction={onAction} />,
              },
            },
            to: `/${account}/${project}/env/${id}`,
          };
        }),
      }}
    />
  );
};

const EnvironmentResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
  const { account } = useOutletContext<IAccountContext>();
  const { project } = useOutletContext<IProjectContext>();
  useWatchReload(
    items.map((i) => {
      return `account:${parseName(account)}.project:${parseName(
        project
      )}.environment:${parseName(i)}`;
    })
  );

  const [visible, setVisible] = useState<BaseType | null>(null);
  const props: IResource = {
    items,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'clone':
          setVisible(item);
          break;
        default:
          break;
      }
    },
  };

  return (
    <>
      <ListGridView
        listView={<ListView {...props} />}
        gridView={<GridView {...props} />}
      />
      <CloneEnvironment
        {...{
          isUpdate: true,
          visible: !!visible,
          setVisible: () => {
            setVisible(null);
          },
          data: visible!,
        }}
      />
    </>
  );
};

export default EnvironmentResourcesV2;
