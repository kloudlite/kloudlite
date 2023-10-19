import { Link } from '@remix-run/react';
import { Badge } from '~/components/atoms/badge';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListItemWithSubtitle,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import { IBuilds } from '~/console/server/gql/queries/build-queries';
import {
  ExtractNodeType,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';

type BaseType = ExtractNodeType<IBuilds>;
const parseItem = (item: BaseType) => {
  return {
    name: item.name,
    id: item.id,
    status: item.status,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

// const ExtraButton = ({ project }: { project: BaseType }) => {
//   const { account } = useParams();
//   return (
//     <ResourceExtraAction
//       options={[
//         {
//           label: 'Settings',
//           icon: <GearSix size={16} />,
//           type: 'item',

//           to: `/${account}/${project.clusterName}/${project.metadata.name}/settings`,
//           key: 'settings',
//         },
//       ]}
//     />
//   );
// };

const GridView = ({ items = [] }: { items: BaseType[] }) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `project-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => <ListTitle title={name} />,
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
  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, status, updateInfo } = parseItem(item);
        const keyPrefix = `project-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, 0),
                className: 'flex-1',
                render: () => <ListTitle title={name} />,
              },
              {
                key: generateKey(keyPrefix, id, index, 'status'),
                className: 'w-[300px]',
                render: () => <Badge>{status}</Badge>,
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
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const BuildResources = ({ items = [] }: { items: BaseType[] }) => {
  return (
    <ListGridView
      listView={<ListView items={items} />}
      gridView={<GridView items={items} />}
    />
  );
};

export default BuildResources;
