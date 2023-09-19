import { DotsThreeVerticalFill, Trash } from '@jengaicons/react';
import { Link } from '@remix-run/react';
import { IconButton } from '~/components/atoms/button';
import { cn, generateKey, titleCase } from '~/components/utils';
import {
  ListBody,
  ListItemWithSubtitle,
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
  return <div className={cn('w-lg h-lg rounded-full', statusColor)} />;
};

const parseItem = (item: ExtractNodeType<IApps>) => {
  return {
    name: item.displayName,
    id: parseName(item),
    path: `/workspaces/${parseName(item)}`,
    updateInfo: {
      author: titleCase(`${parseUpdateOrCreatedBy(item)} updated the app`),
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
        const keyPrefix = `app-${id}-${index}`;
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
                        <AppStatus status={status} />
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
        const keyPrefix = `app-${id}-${index}`;
        return (
          <List.Row
            to={`../app/${id}`}
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'flex-1 bodyMd-semibold text-text-default',
                label: name,
              },
              {
                key: generateKey(keyPrefix, id + name),
                className: 'flex-1 text-text-soft bodyMd',
                label: id,
              },
              {
                key: generateKey(keyPrefix, status),
                className: 'text-text-soft bodyMd w-[200px]',
                render: () => (
                  <div className="flex flex-row gap-lg items-center">
                    <AppStatus status={status} />
                    <span>{uptime}</span>
                  </div>
                ),
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: 'w-[200px]',
                render: () => (
                  <div className="flex flex-col">
                    <div className="text-text-strong bodyMd-medium">
                      {/* Reyan updated the project */}
                      {updateInfo.author}
                    </div>
                    <div className="text-text-soft bodyMd">
                      {updateInfo.time}
                    </div>
                  </div>
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
