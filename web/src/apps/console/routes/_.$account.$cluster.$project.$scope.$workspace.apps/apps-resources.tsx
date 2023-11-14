import { DotsThreeVerticalFill, Trash } from '@jengaicons/react';
import { Link } from '@remix-run/react';
import { IconButton } from '~/components/atoms/button';
import { cn, generateKey, titleCase } from '~/components/utils';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitle,
  ListTitleWithSubtitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { IApps } from '~/console/server/gql/queries/app-queries';
import {
  ExtractNodeType,
  parseName,
  parseStatus,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
  parseUpdateTime,
} from '~/console/server/r-utils/common';

const RESOURCE_NAME = 'app';

const AppStatus = ({ status }: { status: string }) => {
  let statusColor = 'bg-icon-critical';
  switch (status) {
    case 'error':
      statusColor = 'bg-icon-critical';
      break;
    case 'running':
      statusColor = 'bg-icon-success';
      break;
    case 'warning':
      statusColor = 'bg-icon-warning';
      break;
    case 'freezed':
      statusColor = 'bg-icon-soft';
      break;
    default:
      statusColor = 'bg-icon-critical';
  }
  return (
    <div title={status} className={cn('w-lg h-lg rounded-full', statusColor)} />
  );
};

const parseItem = (item: ExtractNodeType<IApps>) => {
  return {
    name: item.displayName,
    id: parseName(item),
    path: `/workspaces/${parseName(item)}`,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
    uptime: parseUpdateTime(item),
    status: parseStatus(item),
  };
};

const GridView = ({ items = [] }: { items: ExtractNodeType<IApps>[] }) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo, uptime, status } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`../app/${id}`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
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
                key: generateKey(keyPrefix, 'status'),
                render: () => (
                  <ListBody
                    data={
                      <div className="flex flex-row gap-lg items-center">
                        <AppStatus status={status.status} />
                        <span>{uptime}</span>
                      </div>
                    }
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

const ListView = ({ items = [] }: { items: ExtractNodeType<IApps>[] }) => {
  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo, uptime, status } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            to={`../app/${id}`}
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'flex-1',
                render: () => <ListTitle title={name} />,
              },
              {
                key: generateKey(keyPrefix, id + name),
                className: 'w-[200px]',
                render: () => <ListBody data={id} />,
              },
              {
                key: generateKey(keyPrefix, status.status),
                className: 'w-[200px]',
                render: () => (
                  <ListBody
                    data={
                      <div className="flex flex-row gap-lg items-center">
                        <AppStatus status={status.status} />
                        <span>{uptime}</span>
                      </div>
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: 'w-[180px]',
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
                  <ResourceExtraAction
                    options={[
                      {
                        key: 'apps-resource-extra-action-1',
                        icon: <Trash size={16} />,
                        label: 'Delete',
                        className: '!text-text-critical',
                        type: 'item',
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

const AppsResources = ({ items = [] }: { items: ExtractNodeType<IApps>[] }) => {
  return (
    <ListGridView
      listView={<ListView items={items} />}
      gridView={<GridView items={items} />}
    />
  );
};

export default AppsResources;
