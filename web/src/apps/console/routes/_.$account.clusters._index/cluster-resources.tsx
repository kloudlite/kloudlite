import { GearSix } from '@jengaicons/react';
import { Link, useParams } from '@remix-run/react';
import { cn, generateKey, titleCase } from '~/components/utils';
import { listRender } from '~/console/components/commons';
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
import { IClusters } from '~/console/server/gql/queries/cluster-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { renderCloudProvider } from '~/console/utils/commons';
import logger from '~/root/lib/client/helpers/log';

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
    case 'do':
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

const ExtraButton = ({ cluster }: { cluster: ExtractNodeType<IClusters> }) => {
  const { account } = useParams();
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Settings',
          icon: <GearSix size={16} />,
          type: 'item',
          to: `/${account}/${cluster.metadata.name}/settings`,
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
        return (
          <Grid.Column
            key={id}
            to={`/${account}/${id}/overview`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={<ExtraButton cluster={item} />}
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
              lR.statusRender({ className: '' }),
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

const ListView = ({ items }: { items: ExtractNodeType<IClusters>[] }) => {
  const { account } = useParams();
  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, provider } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        const lR = listRender({ keyPrefix, resource: item });
        const statusRender = lR.statusRender({
          className: 'w-[180px] mr-[50px]',
        });

        return (
          <List.Row
            key={id}
            className={cn(
              '!p-3xl',
              statusRender.status === 'notready'
                ? '!cursor-default hover:!bg-surface-basic-default'
                : ''
            )}
            // to={`/${account}/${id}/overview`}
            {...(statusRender.status !== 'notready'
              ? { to: `/${account}/${id}/overview` }
              : {})}
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'w-full',
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              statusRender,
              {
                key: generateKey(keyPrefix, `${provider}`),
                className: 'w-[150px] text-start',
                render: () => <ListBody data={provider} />,
              },
              lR.authorRender({ className: 'w-[180px]' }),
              {
                key: generateKey(keyPrefix, 'action'),
                render: () => <ExtraButton cluster={item} />,
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
  return (
    <ListGridView
      gridView={<GridView {...{ items }} />}
      listView={<ListView {...{ items }} />}
    />
  );
};

export default ClusterResources;
