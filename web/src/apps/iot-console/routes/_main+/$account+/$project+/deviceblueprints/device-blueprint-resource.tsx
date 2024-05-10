import { GearSix } from '~/iotconsole/components/icons';
import { Link, useParams } from '@remix-run/react';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/iotconsole/components/console-avatar';
import {
  ListItem,
  ListTitle,
  listClass,
} from '~/iotconsole/components/console-list-components';
import Grid from '~/iotconsole/components/grid';
import ListGridView from '~/iotconsole/components/list-grid-view';
import ResourceExtraAction from '~/iotconsole/components/resource-extra-action';
import {
  ExtractNodeType,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/iotconsole/server/r-utils/common';
import ListV2 from '~/iotconsole/components/listV2';
import { IDeviceBlueprints } from '~/iotconsole/server/gql/queries/iot-device-blueprint-queries';

const RESOURCE_NAME = 'Device blueprint';
type BaseType = ExtractNodeType<IDeviceBlueprints>;

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: item.name,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const ExtraButton = ({ deviceBlueprint }: { deviceBlueprint: BaseType }) => {
  const { account } = useParams();
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Settings',
          icon: <GearSix size={16} />,
          type: 'item',

          to: `/${account}/deviceblueprint/${deviceBlueprint.name}/settings`,
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
            to={`/${account}/deviceblueprint/${id}`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                className: listClass.title,
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={<ExtraButton deviceBlueprint={item} />}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: listClass.author,
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
                render: () => <ExtraButton deviceBlueprint={i} />,
              },
            },
            to: `/${account}/deviceblueprint/${id}`,
          };
        }),
      }}
    />
  );
};

const DeviceBlueprintResource = ({ items = [] }: { items: BaseType[] }) => {
  return (
    <ListGridView
      listView={<ListView items={items} />}
      gridView={<GridView items={items} />}
    />
  );
};

export default DeviceBlueprintResource;
