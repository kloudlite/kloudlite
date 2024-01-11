import { generateKey } from '~/components/utils';
import {
  ListBody,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { IPvcs } from '~/console/server/gql/queries/pvc-queries';

const RESOURCE_NAME = 'storage';
type BaseType = ExtractNodeType<IPvcs>;

const parseItem = (item: BaseType) => {
  return {
    name: parseName(item),
    id: parseName(item),
    updateInfo: {
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

interface IResource {
  items: BaseType[];
}

const GridView = ({ items }: IResource) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => <ListTitle title={name} subtitle={id} />,
              },
              {
                key: generateKey(keyPrefix, 'time'),
                render: () => (
                  <ListBody data={`Last Updated ${updateInfo.time}`} />
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
  return (
    <List.Root>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'flex-1',
                render: () => <ListTitle title={name} subtitle={id} />,
              },
              {
                key: generateKey(keyPrefix, 'time'),
                render: () => (
                  <ListBody data={`Last Updated ${updateInfo.time}`} />
                ),
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const StorageResources = ({ items = [] }: { items: BaseType[] }) => {
  const props: IResource = {
    items,
  };

  return (
    <ListGridView
      listView={<ListView {...props} />}
      gridView={<GridView {...props} />}
    />
  );
};

export default StorageResources;
