import { DotsThreeVerticalFill } from '@jengaicons/react';
import { Link, useParams } from '@remix-run/react';
import { IconButton } from '~/components/atoms/button';
import { Thumbnail } from '~/components/atoms/thumbnail';
import { generateKey, titleCase } from '~/components/utils';
import {
    ListBody,
    ListItem,
    ListItemWithSubtitle,
    ListTitleWithSubtitleAvatar,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import { IEnvironments } from '~/console/server/gql/queries/environment-queries';
import {
    ExtractNodeType,
    parseName,
    parseUpdateOrCreatedBy,
    parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';

const parseItem = (item: ExtractNodeType<IEnvironments>) => {
  return {
    name: item.displayName,
    id: parseName(item),
    cluster: item.clusterName,
    path: `/environments/${parseName(item)}`,
    updateInfo: {
      author: titleCase(
        `${parseUpdateOrCreatedBy(item)} updated the environment`
      ),
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const GridView = ({
  items = [],
}: {
  items: ExtractNodeType<IEnvironments>[];
}) => {
  const { account, project } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, path, cluster, updateInfo } = parseItem(item);
        const keyPrefix = `environment-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/${cluster}/${project}/environment/${id}`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitleWithSubtitleAvatar
                    title={name}
                    subtitle={id}
                    action={
                      <IconButton
                        icon={<DotsThreeVerticalFill />}
                        variant="plain"
                        onClick={(e) => e.stopPropagation()}
                      />
                    }
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
                key: generateKey(keyPrefix, path + cluster),
                render: () => (
                  <div className="flex flex-col gap-md">
                    <ListItem data={path} />
                    <ListBody data={cluster} />
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

const ListView = ({ items }: { items: ExtractNodeType<IEnvironments>[] }) => {
  const { account, project } = useParams();

  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, path, cluster, updateInfo } = parseItem(item);
        const keyPrefix = `envrionment-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            to={`/${account}/${cluster}/${project}/environment/${id}`}
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
                key: generateKey(keyPrefix, cluster),
                className: 'w-[120px] text-start',
                render: () => <ListBody data={cluster} />,
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
                render: () => (
                  <IconButton
                    icon={<DotsThreeVerticalFill />}
                    variant="plain"
                    onClick={(e) => e.stopPropagation()}
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

const Resources = ({
  items = [],
}: {
  items: ExtractNodeType<IEnvironments>[];
}) => {
  return (
    <ListGridView
      listView={<ListView items={items} />}
      gridView={<GridView items={items} />}
    />
  );
};

export default Resources;
