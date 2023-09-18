import { DotsThreeVerticalFill } from '@jengaicons/react';
import { Link, useParams } from '@remix-run/react';
import { IconButton } from '~/components/atoms/button';
import { Thumbnail } from '~/components/atoms/thumbnail';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithSubtitleAvatar,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import { IProjects } from '~/console/server/gql/queries/project-queries';
import { ExtractNodeType } from '~/console/server/r-utils/common';

const parseItem = (item: any) => {
  return {
    name: item.name,
    id: item.id,
    type: item.type,
    updateInfo: item.updateInfo,
  };
};

const genKey = (...items: Array<string | number>) => items.join('-');

const GridView = ({ items = [] }: { items: any }) => {
  const { account } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, type, updateInfo } = parseItem(item);

        return (
          <Grid.Column
            key={id}
            rows={[
              {
                key: genKey('backend-services', id, index, 0),
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
                key: genKey('backend-services', id, index, 1),
                render: () => (
                  <div className="flex flex-col gap-md">
                    <ListBody data={type} />
                  </div>
                ),
              },
              {
                key: genKey('backend-services', id, index, 2),
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
        const { name, id, type, updateInfo } = parseItem(item);

        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: genKey('backend-services', id, index, 0),
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
                key: genKey('backend-services', id, index, 3),
                className: 'w-[120px] text-start',
                render: () => <ListBody data={type} />,
              },
              {
                key: genKey('backend-services', id, index, 4),
                render: () => (
                  <ListItemWithSubtitle
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              {
                key: genKey('backend-services', id, index, 5),
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

const BackendServicesResources = ({ items = [] }: { items: any }) => {
  return (
    <ListGridView
      listView={<ListView items={items} />}
      gridView={<GridView items={items} />}
    />
  );
};

export default BackendServicesResources;
