import { DotsThreeVerticalFill } from '@jengaicons/react';
import { Link, useParams } from '@remix-run/react';
import { IconButton } from '~/components/atoms/button';
import { Thumbnail } from '~/components/atoms/thumbnail';
import { dayjs } from '~/components/molecule/dayjs';
import { titleCase } from '~/components/utils';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithSubtitle,
  ListTitleWithSubtitleAvatar,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import { ICluster } from '~/console/server/gql/queries/cluster-queries';
import { parseFromAnn, parseName } from '~/console/server/r-utils/common';
import { keyconstants } from '~/console/server/r-utils/key-constants';

const parseItem = (item: ICluster) => {
  return {
    name: item.displayName,
    id: parseName(item),
    path: `/clusters/${parseName(item)}`,
    provider: `${item?.spec?.cloudProvider} (${item?.spec?.region})` || '',
    updateInfo: {
      author: titleCase(
        `${parseFromAnn(item, keyconstants.author)} updated the project`
      ),
      time: dayjs(item.updateTime).fromNow(),
    },
  };
};

const genKey = (...items: Array<string | number>) => items.join('-');

const GridView = ({ items = [] }: { items: ICluster[] }) => {
  const { account } = useParams();
  return (
    <Grid.Root className="!grid-cols-3">
      {items.map((item, index) => {
        const { name, id, path, provider, updateInfo } = parseItem(item);

        return (
          <Grid.Column
            key={id}
            rows={[
              {
                key: genKey('cluster', id, index, 0),
                render: () => (
                  <ListTitleWithSubtitle
                    title={name}
                    subtitle={id}
                    action={
                      <IconButton
                        icon={<DotsThreeVerticalFill />}
                        variant="plain"
                        onClick={(e) => e.stopPropagation()}
                      />
                    }
                  />
                ),
              },
              {
                key: genKey('cluster', id, index, 1),
                render: () => (
                  <div className="flex flex-col gap-md">
                    <ListBody data={path} />
                    <ListBody data={provider} />
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

const ListView = ({ items = [] }: { items: ICluster[] }) => {
  const { account } = useParams();
  return (
    <List.Root linkComponent={Link}>
      {items.map((item) => {
        const { name, id, path, provider, updateInfo } = parseItem(item);
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            to={`/${account}/${id}/nodepools`}
            columns={[
              {
                key: 1,
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
                key: 2,
                className: 'w-[230px] text-start',
                render: () => <ListBody data={path} />,
              },
              {
                key: 3,
                className: 'w-[120px] text-start',
                render: () => <ListBody data={provider} />,
              },
              {
                key: 4,
                render: () => (
                  <ListItemWithSubtitle
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              {
                key: 5,
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

const Resources = ({ items = [] }: { items: ICluster[] }) => {
  return <ListView items={items} />;
};

export default Resources;
