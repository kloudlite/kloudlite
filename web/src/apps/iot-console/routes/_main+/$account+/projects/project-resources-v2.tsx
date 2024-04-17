import { GearSix } from '~/iotconsole/components/icons';
import { Link, useOutletContext, useParams } from '@remix-run/react';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/iotconsole/components/console-avatar';
import {
  ListItem,
  ListTitle,
} from '~/iotconsole/components/console-list-components';
import Grid from '~/iotconsole/components/grid';
import ListGridView from '~/iotconsole/components/list-grid-view';
import ResourceExtraAction from '~/iotconsole/components/resource-extra-action';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/iotconsole/server/r-utils/common';
import { useWatchReload } from '~/root/lib/client/helpers/socket/useWatch';
import ListV2 from '~/iotconsole/components/listV2';
import { IProjects } from '~/iotconsole/server/gql/queries/iot-project-queries';
import { IAccountContext } from '../_layout';

type BaseType = ExtractNodeType<IProjects>;
const RESOURCE_NAME = 'project';

const parseItem = (item: ExtractNodeType<IProjects>) => {
  return {
    name: item.displayName,
    id: item.name,
    path: `/projects/${item.name}`,
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

          to: `/${account}/${project.name}/settings`,
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
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/${id}/deviceblueprints`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={<ExtraButton project={item} />}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
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
            className: 'w-[180px] flex-1',
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
          const { name, id, updateInfo } = parseItem(i);
          console.log('updateInfo', parseItem(i));
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
              updated: {
                render: () => (
                  <ListItem
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              action: {
                render: () => <ExtraButton project={i} />,
              },
            },
            to: `/${account}/${id}/deviceblueprints`,
          };
        }),
      }}
    />
  );
};

const ProjectResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
  const { account } = useOutletContext<IAccountContext>();
  useWatchReload(`account:${parseName(account)}`);
  return (
    <ListGridView
      listView={<ListView items={items} />}
      gridView={<GridView items={items} />}
    />
  );
};

export default ProjectResourcesV2;
