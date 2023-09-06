import { DotsThreeVerticalFill, Trash } from '@jengaicons/react';
import { IconButton } from '~/components/atoms/button';
import { cn } from '~/components/utils';
import List from '~/console/components/list';
import ResourceExtraAction from '~/console/components/resource-extra-action';

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

const AppsResources = ({ items = [] }: { items: Array<any> }) => {
  return (
    <List.Root>
      {items.map((item: any) => (
        <List.Row
          key={item.id}
          className="!p-3xl"
          columns={[
            {
              key: 1,
              className: 'flex-1 bodyMd-semibold text-text-default',
              label: item.name,
            },
            {
              key: 2,
              className: 'flex-1 text-text-soft bodyMd',
              label: item.id,
            },
            {
              key: 3,
              className: 'text-text-soft bodyMd w-[200px]',
              render: () => (
                <div className="flex flex-row gap-lg items-center">
                  <AppStatus status={item.status} />
                  <span>{item.uptime}</span>
                </div>
              ),
            },
            {
              key: 4,
              className: 'w-[200px]',
              render: () => (
                <div className="flex flex-col">
                  <div className="text-text-strong bodyMd-medium">
                    Reyan updated the project
                  </div>
                  <div className="text-text-soft bodyMd">3 days ago</div>
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
