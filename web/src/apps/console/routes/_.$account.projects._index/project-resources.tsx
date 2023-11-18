import { GearSix } from '@jengaicons/react';
import { Link, useParams } from '@remix-run/react';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithSubtitleAvatar,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { IProjects } from '~/console/server/gql/queries/project-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';

type BaseType = ExtractNodeType<IProjects>;
const RESOURCE_NAME = 'project';

const parseItem = (item: ExtractNodeType<IProjects>) => {
  return {
    name: item.displayName,
    id: parseName(item),
    cluster: item.clusterName,
    path: `/projects/${parseName(item)}`,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const ExtraButton = ({ project }: { project: BaseType }) => {
  const { account } = useParams();
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Settings',
          icon: <GearSix size={16} />,
          type: 'item',

          to: `/${account}/${project.clusterName}/${project.metadata.name}/settings`,
          key: 'settings',
        },
      ]}
    />
  );
};

const GridView = ({ items = [] }: { items: BaseType[] }) => {
  const { account } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, path, cluster, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/${cluster}/${id}/workspaces`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitleWithSubtitleAvatar
                    title={name}
                    subtitle={id}
                    action={<ExtraButton project={item} />}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, path + cluster),
                render: () => (
                  <div className="flex flex-col gap-md">
                    {/* <ListItem data={path} /> */}
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

const ListView = ({ items }: { items: BaseType[] }) => {
  const { account } = useParams();
  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, cluster, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            to={`/${account}/${cluster}/${id}/workspaces`}
            columns={[
              {
                key: generateKey(keyPrefix, 0),
                className: 'flex-1',
                render: () => (
                  <ListTitleWithSubtitleAvatar
                    title={name}
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              // {
              //   key: generateKey(keyPrefix, path),
              //   className: 'w-[230px] text-start',
              //   render: () => <ListBody data={path} />,
              // },
              {
                key: generateKey(keyPrefix, cluster),
                className: 'w-[120px] text-start',
                render: () => <ListBody data={cluster} />,
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
                render: () => <ExtraButton project={item} />,
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const ProjectResources = ({ items = [] }: { items: BaseType[] }) => {
  return (
    <ListGridView
      listView={<ListView items={items} />}
      gridView={<GridView items={items} />}
    />
  );
};

export default ProjectResources;
