import { Trash } from '@jengaicons/react';
import { cn } from '~/components/utils';
import List from '~/console/components/list';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { IApps } from '~/console/server/gql/queries/app-queries';
import {
  parseStatus,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
  parseUpdateTime,
  ExtractNodeType,
  parseName,
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

const AppsResources = ({ items = [] }: { items: ExtractNodeType<IApps>[] }) => {
  return (
    <List.Root>
      {items.map((item) => (
        <List.Row
          key={parseName(item)}
          className="!p-3xl"
          columns={[
            {
              key: 1,
              className: 'flex-1 bodyMd-semibold text-text-default',
              label: item.displayName,
            },
            {
              key: 2,
              className: 'flex-1 text-text-soft bodyMd',
              label: parseName(item),
            },
            {
              key: 3,
              className: 'text-text-soft bodyMd w-[200px]',
              render: () => (
                <div className="flex flex-row gap-lg items-center">
                  <AppStatus status={parseStatus(item)} />
                  <span>{parseUpdateTime(item)}</span>
                </div>
              ),
            },
            {
              key: 4,
              className: 'w-[200px]',
              render: () => (
                <div className="flex flex-col">
                  <div className="text-text-strong bodyMd-medium">
                    {/* Reyan updated the project */}
                    {parseUpdateOrCreatedBy(item)}
                  </div>
                  <div className="text-text-soft bodyMd">
                    {parseUpdateOrCreatedOn(item)}
                  </div>
                </div>
              ),
            },
            {
              key: 5,
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
      ))}
    </List.Root>
  );
};

export default AppsResources;
