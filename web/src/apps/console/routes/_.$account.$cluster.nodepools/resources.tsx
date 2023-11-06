import { PencilLine } from '@jengaicons/react';
import { Link, useParams } from '@remix-run/react';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListItemWithSubtitle,
  ListTitleWithSubtitleAvatar,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { INodepools } from '~/console/server/gql/queries/nodepool-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';

const parseItem = (item: ExtractNodeType<INodepools>) => {
  return {
    name: item.displayName,
    id: parseName(item),
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const RESOURCE_NAME = 'nodepool';
type BaseType = ExtractNodeType<INodepools>;

interface IResource {
  items: BaseType[];
  onEdit: (item: ExtractNodeType<INodepools>) => void;
}

const GridView = ({ items, onEdit }: IResource) => {
  const { account } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            onClick={() => {}}
            key={id}
            to={`/${account}/${id}/nodepools`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ResourceExtraAction
                    options={[
                      {
                        key: '1',
                        label: 'Edit',
                        icon: <PencilLine size={16} />,
                        type: 'item',
                        onClick: () => onEdit(item),
                      },
                    ]}
                  />
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

const ListView = ({ items, onEdit }: IResource) => {
  const { account } = useParams();
  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
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
                render: () => (
                  <ResourceExtraAction
                    options={[
                      {
                        key: '1',
                        label: 'Edit',
                        icon: <PencilLine size={16} />,
                        type: 'item',
                        onClick: () => onEdit(item),
                      },
                    ]}
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

const NodepoolResources = ({
  items = [],
  onEdit,
}: {
  items: BaseType[];
  onEdit: (item: ExtractNodeType<INodepools>) => void;
}) => {
  const props: IResource = {
    items,
    onEdit,
  };

  return (
    <ListGridView
      gridView={<GridView {...props} />}
      listView={<ListView {...props} />}
    />
  );
};

export default NodepoolResources;
