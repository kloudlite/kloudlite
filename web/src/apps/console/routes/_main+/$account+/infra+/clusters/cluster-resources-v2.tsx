import { GearSix } from '@jengaicons/react';
import { Link, useOutletContext, useParams } from '@remix-run/react';
import { generateKey, titleCase } from '~/components/utils';
import { listRender } from '~/console/components/commons';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListBody,
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { SyncStatusV2 } from '~/console/components/sync-status';
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
import ListV2 from '~/console/components/listV2';
import AnimateHide from '~/components/atoms/animate-hide';
import LogComp from '~/root/lib/client/components/logger';
import LogAction from '~/console/page-components/log-action';
import { Button } from '~/components/atoms/button';
import { useDataState } from '~/console/page-components/common-state';
import { useState } from 'react';
import { dayjs } from '~/components/molecule/dayjs';

type BaseType = ExtractNodeType<IClusters>;
const RESOURCE_NAME = 'cluster';

const getProvider = (item: BaseType) => {
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

const parseItem = (item: BaseType) => {
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

const ExtraButton = ({ cluster }: { cluster: BaseType }) => {
  const { account } = useParams();
  return (
    <ResourceExtraAction
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

const GridView = ({ items }: { items: BaseType[] }) => {
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
                      // <ExtraButton status={status.status} cluster={item} />
                      <span />
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
const ListView = ({ items }: { items: BaseType[] }) => {
  const [open, setOpen] = useState<string>('');
  const { state } = useDataState<{
    linesVisible: boolean;
    timestampVisible: boolean;
  }>('logs');

  const { account } = useParams();
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
          // {
          //   render: () => '',
          //   name: 'logs',
          //   className: 'w-[180px]',
          // },
          {
            render: () => '',
            name: 'status',
            className: 'flex-1 min-w-[30px] flex items-center justify-center',
          },
          {
            render: () => 'Provider (Region)',
            name: 'provider',
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
          const { name, id, updateInfo, provider } = parseItem(i);

          // const isLatest = dayjs(i.updateTime).isAfter(
          //   dayjs().subtract(3, 'hour')
          // );

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
              // logs: {
              //   render: () => (
              //     <Button
              //       size="sm"
              //       variant="basic"
              //       content={open === i.id ? 'Hide Logs' : 'Show Logs'}
              //       onClick={(e) => {
              //         e.preventDefault();

              //         setOpen((s) => {
              //           if (s === i.id) {
              //             return '';
              //           }
              //           return i.id;
              //         });
              //       }}
              //     />
              //   ),
              // },
              status: {
                render: () => null,
              },
              provider: { render: () => <ListItem data={provider} /> },
              updated: {
                render: () => (
                  <ListItem
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              action: {
                render: () => <ExtraButton cluster={i} />,
              },
            },
            to: `/${account}/infra/${id}/overview`,
            detail: (
              <AnimateHide
                onClick={(e) => e.preventDefault()}
                show={open === i.id}
                className="w-full flex pt-4xl pb-2xl justify-center items-center"
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
                      account: account || '',
                      cluster: parseName(i),
                      trackingId: i.id,
                    },
                    actionComponent: <LogAction />,
                  }}
                />
              </AnimateHide>
            ),
            hideDetailSeperator: true,
          };
        }),
      }}
    />
  );
};

const ClusterResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
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

export default ClusterResourcesV2;
