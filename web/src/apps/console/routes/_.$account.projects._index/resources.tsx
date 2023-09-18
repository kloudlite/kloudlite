import { DotsThreeVerticalFill } from '@jengaicons/react';
import { Link, useParams } from '@remix-run/react';
import { IconButton } from '~/components/atoms/button';
import { Thumbnail } from '~/components/atoms/thumbnail';
import { dayjs } from '~/components/molecule/dayjs';
import { titleCase } from '~/components/utils';
import {
  ListBody,
  ListItem,
  ListItemWithSubtitle,
  ListTitleWithSubtitleAvatar,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import { IProjects } from '~/console/server/gql/queries/project-queries';
import {
  ExtractNodeType,
  parseFromAnn,
  parseName,
} from '~/console/server/r-utils/common';
import { keyconstants } from '~/console/server/r-utils/key-constants';

const parseItem = (item: ExtractNodeType<IProjects>) => {
  return {
    name: item.displayName,
    id: parseName(item),
    cluster: item.clusterName,
    path: `/projects/${parseName(item)}`,
    updateInfo: {
      author: titleCase(
        `${parseFromAnn(item, keyconstants.author)} updated the project`
      ),
      time: dayjs(item.updateTime).fromNow(),
    },
  };
};

const genKey = (...items: Array<string | number>) => items.join('-');

const GridView = ({ items = [] }: { items: ExtractNodeType<IProjects>[] }) => {
  const { account } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, path, cluster, updateInfo } = parseItem(item);

        return (
          <Grid.Column
            key={id}
            to={`/${account}/${cluster}/${id}/workspaces`}
            rows={[
              {
                key: genKey('project', id, index, 0),
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
                key: genKey('cluster', id, index, 1),
                render: () => (
                  <div className="flex flex-col gap-md">
                    <ListItem data={path} />
                    <ListBody data={cluster} />
                  </div>
                ),
              },
              {
                key: genKey('cluster', id, index, 2),
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

const ListView = ({ items }: { items: ExtractNodeType<IProjects>[] }) => {
  const { account } = useParams();
  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, cluster, path, updateInfo } = parseItem(item);

        return (
          <List.Row
            key={id}
            className="!p-3xl"
            to={`/${account}/${cluster}/${id}/workspaces`}
            columns={[
              {
                key: genKey('project', id, index, 0),
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
                key: genKey('project', id, index, 1),
                className: 'w-[230px] text-start',
                render: () => <ListBody data={path} />,
              },
              {
                key: genKey('project', id, index, 3),
                className: 'w-[120px] text-start',
                render: () => <ListBody data={cluster} />,
              },
              {
                key: genKey('project', id, index, 4),
                render: () => (
                  <ListItemWithSubtitle
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              {
                key: genKey('project', id, index, 5),
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

const Resources = ({ items = [] }: { items: ExtractNodeType<IProjects>[] }) => {
  return (
    <ListGridView
      listView={<ListView items={items} />}
      gridView={<GridView items={items} />}
    />
  );
};

export default Resources;
