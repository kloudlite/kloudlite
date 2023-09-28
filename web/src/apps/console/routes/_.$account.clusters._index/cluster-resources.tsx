import { GearSix } from '@jengaicons/react';
import { Link, useParams } from '@remix-run/react';
import { Thumbnail } from '~/components/atoms/thumbnail';
import { dayjs } from '~/components/molecule/dayjs';
import { generateKey, titleCase } from '~/components/utils';
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

const RESOURCE_NAME = 'cluster';
type BaseType = ExtractNodeType<IClusters>;

interface IResource {
  items: BaseType[];
}

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: parseName(item),
    path: `/clusters/${parseName(item)}`,
    provider: `${item?.spec?.cloudProvider} (${item?.spec?.region})` || '',
    updateInfo: {
      author: titleCase(
        `${parseFromAnn(
          item,
          keyconstants.author
        )} updated the ${RESOURCE_NAME}`
      ),
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
        const { name, id, path, provider, updateInfo } = parseItem(item);
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
                    avatar={
                      <Thumbnail
                        size="sm"
                        rounded
                        src="https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"
                      />
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, path),
                className: 'w-[230px] text-start',
                render: () => <ListBody data={path} />,
              },
              {
                key: generateKey(keyPrefix, provider),
                className: 'w-[120px] text-start',
                render: () => <ListBody data={provider} />,
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
