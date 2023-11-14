import { GearSix } from '@jengaicons/react';
import { Link, useParams } from '@remix-run/react';
import { dayjs } from '~/components/molecule/dayjs';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithSubtitle,
  ListTitleWithSubtitleAvatar,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { IClusters } from '~/console/server/gql/queries/cluster-queries';
import {
  ExtractNodeType,
  parseFromAnn,
  parseName,
} from '~/console/server/r-utils/common';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import logger from '~/root/lib/client/helpers/log';
import { Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider as IClusterSpecCloudProvider } from '~/root/src/generated/gql/server';

const RESOURCE_NAME = 'cluster';
type BaseType = ExtractNodeType<IClusters>;

interface IResource {
  items: BaseType[];
}

const getProvider = (item: ExtractNodeType<IClusters>) => {
  if (!item.spec) {
    return '';
  }
  switch (item.spec.cloudProvider as IClusterSpecCloudProvider) {
    case 'aws':
      return `${item.spec.cloudProvider} (${item.spec.aws?.region})`;

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
      author: `Updated by ${titleCase(
        parseFromAnn(item, keyconstants.author)
      )}`,
      time: dayjs(item.updateTime).fromNow(),
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

const GridView = ({ items }: IResource) => {
  const { account } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, provider, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/${id}/nodepools`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitleWithSubtitle
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
              {
                key: generateKey(keyPrefix, updateInfo.author),
                render: () => (
                  <ListItemWithSubtitle
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

const ListView = ({ items }: IResource) => {
  const { account } = useParams();
  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, provider, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            to={`/${account}/${id}/nodepools`}
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'flex-1',
                render: () => (
                  <ListTitleWithSubtitleAvatar
                    title={name}
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, provider),
                className: 'w-[150px] text-start',
                render: () => <ListBody data={provider} />,
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: 'w-[180px]',
                render: () => (
                  <ListItemWithSubtitle
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
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

const ClusterResources = ({ items = [] }: { items: BaseType[] }) => {
  const props: IResource = {
    items,
  };

  return (
    <ListGridView
      gridView={<GridView {...props} />}
      listView={<ListView {...props} />}
    />
  );
};

export default ClusterResources;
