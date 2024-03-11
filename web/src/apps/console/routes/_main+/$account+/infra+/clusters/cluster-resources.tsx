import { GearSix } from '@jengaicons/react';
import { Link, useOutletContext, useParams } from '@remix-run/react';
import { generateKey, titleCase } from '~/components/utils';
import { IStatus, listRender } from '~/console/components/commons';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListBody,
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { listStatus } from '~/console/components/sync-status';
import { IClusters } from '~/console/server/gql/queries/cluster-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { renderCloudProvider } from '~/console/utils/commons';
import logger from '~/root/lib/client/helpers/log';
import { IAccountContext } from '~/console/routes/_main+/$account+/_layout';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import { ISetState } from '~/console/page-components/app-states';
import { useState } from 'react';
import { dayjs } from '~/components/molecule/dayjs';
import { Button } from '~/components/atoms/button';
import AnimateHide from '~/components/atoms/animate-hide';
import LogComp from '~/root/lib/client/components/logger';
import LogAction from '~/console/page-components/log-action';
import { useDataState } from '~/console/page-components/common-state';

const RESOURCE_NAME = 'cluster';

const getProvider = (item: ExtractNodeType<IClusters>) => {
  if (!item.spec) {
    return '';
  }
  switch (item.spec.cloudProvider) {
    case 'aws':
      return (
        <div className="flex flex-row items-center gap-lg">
          {renderCloudProvider({ cloudprovider: item.spec.cloudProvider })}
          <span>({item.spec.aws?.region})</span>
        </div>
      );
    case 'gcp':
    case 'azure':
      return (
        <div className="flex flex-row items-center gap-lg">
          <span>{item.spec.cloudProvider}</span>
        </div>
      );

    default:
      logger.error('unknown provider', item.spec.cloudProvider);
      return '';
  }
};

const parseItem = (item: ExtractNodeType<IClusters>) => {
  return {
    name: item.displayName,
    id: parseName(item),
    provider: getProvider(item),
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const ExtraButton = ({
  cluster,
  status,
}: {
  cluster: ExtractNodeType<IClusters>;
  status: IStatus;
}) => {
  const { account } = useParams();
  return (
    <ResourceExtraAction
      disabled={status === 'deleting' || status === 'syncing'}
      options={[
        {
          label: 'Settings',
          icon: <GearSix size={16} />,
          type: 'item',
          to: `/${account}/infra/${cluster.metadata.name}/settings`,
          key: 'settings',
        },
      ]}
    />
  );
};

const GridView = ({ items }: { items: ExtractNodeType<IClusters>[] }) => {
  const { account } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, provider, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        const lR = listRender({ keyPrefix, resource: item });
        const status = lR.statusRender({ className: '' });
        return (
          <Grid.Column
            key={id}
            to={`/${account}/infra/${id}/overview`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={
                      <ExtraButton status={status.status} cluster={item} />
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, id + name + provider),
                render: () => (
                  <div className="flex flex-col gap-md">
                    {/* <ListItem data={path} /> */}
                    <ListBody data={provider} />
                  </div>
                ),
              },
              status,
              {
                key: generateKey(keyPrefix, updateInfo.author),
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

type BaseType = ExtractNodeType<IClusters>;

interface IResource {
  items: BaseType[];
  // onDelete: (item: BaseType) => void;
  // onEdit: (item: BaseType) => void;
}

const ListDetail = (
  props: Omit<IResource, 'items'> & {
    open: string;
    item: BaseType;
    setOpen: ISetState<string>;
  }
) => {
  const { item, open, setOpen } = props;

  const { name, id } = parseItem(item);
  const keyPrefix = `${RESOURCE_NAME}-${id}`;
  const lR = listRender({ keyPrefix, resource: item });

  const { account } = useOutletContext<IAccountContext>();

  const isLatest = dayjs(item.updateTime).isAfter(dayjs().subtract(3, 'hour'));

  const tempStatus = listStatus({
    key: keyPrefix,
    item,
    className: 'basis-full text-center',
  });

  const { state } = useDataState<{
    linesVisible: boolean;
    timestampVisible: boolean;
  }>('logs');

  return (
    <div className="w-full flex flex-col">
      <div className="flex flex-row items-center">
        <div className="w-[220px] min-w-[220px]  mr-xl flex flex-row items-center">
          <ListTitle
            title={name}
            subtitle={
              <div className="flex flex-row items-center gap-md">{id}</div>
            }
            avatar={<ConsoleAvatar name={id} />}
          />
        </div>

        {isLatest && (
          <Button
            size="sm"
            variant="basic"
            content={open === item.id ? 'Hide Logs' : 'Show Logs'}
            onClick={(e) => {
              e.preventDefault();

              setOpen((s) => {
                if (s === item.id) {
                  return '';
                }
                return item.id;
              });
            }}
          />
        )}

        <div className="flex items-center w-[20px] mx-xl flex-grow">
          {tempStatus.render()}
        </div>

        <div className="pr-3xl w-[180px] min-w-[180px]">
          {lR.authorRender({ className: '' }).render()}
        </div>
      </div>

      <AnimateHide
        onClick={(e) => e.preventDefault()}
        show={open === item.id}
        className="w-full flex pt-4xl justify-center items-center"
      >
        <LogComp
          {...{
            hideLineNumber: !state.linesVisible,
            hideTimestamp: !state.timestampVisible,
            className: 'flex-1',
            dark: true,
            width: '100%',
            height: '40rem',
            title: 'Logs',
            websocket: {
              account: parseName(account),
              cluster: parseName(item),
              trackingId: item.id,
            },
            actionComponent: <LogAction />,
          }}
        />
      </AnimateHide>
    </div>
  );
};

const ListView = ({ items }: IResource) => {
  const [open, setOpen] = useState<string>('');
  const { account } = useParams();

  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;

        const lR = listRender({ keyPrefix, resource: item });
        const statusRender = lR.statusRender({
          className: 'min-w-[80px] mx-[25px] basis-full text-center',
        });

        return (
          <List.Row
            key={id}
            className="!p-3xl"
            {...(!(
              statusRender.status === 'notready' ||
              statusRender.status === 'deleting'
            )
              ? { to: `/${account}/infra/${id}/overview` }
              : {})}
            columns={[
              {
                className: 'w-full',
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListDetail item={item} open={open} setOpen={setOpen} />
                ),
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const ClusterResources = ({
  items = [],
}: {
  items: ExtractNodeType<IClusters>[];
}) => {
  const { account } = useOutletContext<IAccountContext>();
  useWatchReload(
    items.map((i) => {
      return `account:${parseName(account)}.cluster:${parseName(i)}`;
    })
  );

  return (
    <ListGridView
      gridView={<GridView {...{ items }} />}
      listView={<ListView {...{ items }} />}
    />
  );
};

export default ClusterResources;
