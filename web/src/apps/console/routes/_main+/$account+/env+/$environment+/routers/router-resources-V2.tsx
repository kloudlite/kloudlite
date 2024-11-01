import { Trash, PencilLine } from '~/console/components/icons';
import { useState } from 'react';
import { generateKey, titleCase } from '@kloudlite/design-system/utils';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { handleError } from '~/lib/utils/common';
import { IRouters } from '~/console/server/gql/queries/router-queries';
import { Link, useParams } from '@remix-run/react';
import { SyncStatusV2 } from '~/console/components/sync-status';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/lib/client/helpers/reloader';
import { toast } from '@kloudlite/design-system/molecule/toast';
import { Button } from '@kloudlite/design-system/atoms/button';
import Tooltip from '@kloudlite/design-system/atoms/tooltip';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import ListV2 from '~/console/components/listV2';
import HandleRouter from './handle-router';

const RESOURCE_NAME = 'domain';
type BaseType = ExtractNodeType<IRouters>;

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: parseName(item),
    updateInfo: {
      author: `Updated by ${parseUpdateOrCreatedBy(item)}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

type OnAction = ({
  action,
  item,
}: {
  action: 'edit' | 'delete' | 'detail';
  item: BaseType;
}) => void;

type IExtraButton = {
  onAction: OnAction;
  item: BaseType;
};

const formatDomain = (domain: string) => {
  const d = domain.startsWith('https://') ? domain : `https://${domain}`;
  return { full: d, short: d.replace('https://', '') };
};

const ExtraButton = ({ onAction, item }: IExtraButton) => {
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Edit',
          icon: <PencilLine size={16} />,
          type: 'item',
          onClick: () => onAction({ action: 'edit', item }),
          key: 'edit',
        },
        {
          label: 'Delete',
          icon: <Trash size={16} />,
          type: 'item',
          onClick: () => onAction({ action: 'delete', item }),
          key: 'delete',
          className: '!text-text-critical',
        },
      ]}
    />
  );
};

interface IResource {
  items: BaseType[];
  onAction: OnAction;
}

const GridView = ({ items, onAction }: IResource) => {
  const { account, environment } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        const firstDomain = item.spec.domains?.[0];
        return (
          <Grid.Column
            key={id}
            to={`/${account}/env/${environment}/router/${id}/routes`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={name}
                    action={<ExtraButton onAction={onAction} item={item} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'extra_domain'),
                render: () => (
                  <ListItem
                    data={
                      <div className="flex flex-row items-center gap-md">
                        <Button
                          linkComponent={Link}
                          target="_blank"
                          size="sm"
                          content={formatDomain(firstDomain).short}
                          variant="primary-plain"
                          to={formatDomain(firstDomain).full}
                        />

                        {item.spec.domains.length > 1 && (
                          <Tooltip.Root
                            content={
                              <div className="flex flex-col gap-md">
                                {item.spec.domains
                                  .filter((d) => d !== firstDomain)
                                  .map((d) => (
                                    <Button
                                      key={d}
                                      linkComponent={Link}
                                      target="_blank"
                                      size="sm"
                                      content={formatDomain(d).short}
                                      variant="primary-plain"
                                      to={formatDomain(d).full}
                                    />
                                  ))}
                              </div>
                            }
                          >
                            <Button
                              content={`+${item.spec.domains.length - 1} more`}
                              variant="plain"
                              size="sm"
                            />
                          </Tooltip.Root>
                        )}
                      </div>
                    }
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

const ListView = ({ items, onAction }: IResource) => {
  const { account, environment } = useParams();
  return (
    <ListV2.Root
      linkComponent={Link}
      data={{
        headers: [
          {
            render: () => 'Name',
            name: 'name',
            className: 'w-[180px] shrink-0',
          },
          {
            render: () => 'Status',
            name: 'status',
            className:
              'flex-1 min-w-[80px] flex items-center lg:justify-center shrink-0',
          },
          {
            render: () => 'Domains',
            name: 'domains',
            className: 'min-w-[280px] max-w-[280px] lg:max-w-none lg:flex-1',
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
          const firstDomain = i.spec.domains?.[0];
          return {
            to: `/${account}/env/${environment}/router/${id}/routes`,
            columns: {
              name: {
                render: () => <ListTitle title={name} />,
              },
              status: {
                render: () => <SyncStatusV2 item={i} />,
              },
              domains: {
                render: () => (
                  <div className="flex flex-row items-center gap-md">
                    <Tooltip.Root
                      content={
                        <Button
                          size="sm"
                          content={
                            <span className="text-left">
                              {formatDomain(firstDomain).short}
                            </span>
                          }
                          variant="primary-plain"
                          className="!pl-0"
                          onClick={(e) => {
                            e.preventDefault();
                            window.open(
                              formatDomain(firstDomain).full,
                              '_blank',
                              'noopener,noreferrer'
                            );
                          }}
                        />
                      }
                    >
                      <Button
                        size="sm"
                        content={
                          <span className="truncate text-left">
                            {formatDomain(firstDomain).short}
                          </span>
                        }
                        variant="primary-plain"
                        className="truncate"
                        onClick={(e) => {
                          e.preventDefault();
                          window.open(
                            formatDomain(firstDomain).full,
                            '_blank',
                            'noopener,noreferrer'
                          );
                        }}
                      />
                    </Tooltip.Root>
                    {i.spec.domains.length > 1 && (
                      <Tooltip.Root
                        content={
                          <div className="flex flex-col gap-md">
                            {i.spec.domains
                              .filter((d) => d !== firstDomain)
                              .map((d) => (
                                <Button
                                  key={d}
                                  size="sm"
                                  content={
                                    <span className="text-left">
                                      {formatDomain(d).short}
                                    </span>
                                  }
                                  variant="primary-plain"
                                  onClick={(e) => {
                                    e.preventDefault();
                                    window.open(
                                      formatDomain(d).full,
                                      '_blank',
                                      'noopener,noreferrer'
                                    );
                                  }}
                                />
                              ))}
                          </div>
                        }
                      >
                        <div className="shrink-0">
                          <Button
                            content={`+${i.spec.domains.length - 1} more`}
                            variant="plain"
                            size="sm"
                          />
                        </div>
                      </Tooltip.Root>
                    )}
                  </div>
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
                render: () => <ExtraButton item={i} onAction={onAction} />,
              },
            },
          };
        }),
      }}
    />
  );
};

const RouterResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [visible, setVisible] = useState<BaseType | null>(null);
  const api = useConsoleApi();
  const reloadPage = useReload();
  const { environment, account } = useParams();

  useWatchReload(
    items.map((i) => {
      return `account:${account}.environment:${environment}.router:${parseName(
        i
      )}`;
    })
  );

  const props: IResource = {
    items,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'edit':
          setVisible(item);
          break;
        case 'delete':
          setShowDeleteDialog(item);
          break;
        default:
      }
    },
  };
  return (
    <>
      <ListGridView
        listView={<ListView {...props} />}
        gridView={<GridView {...props} />}
      />
      <DeleteDialog
        resourceName={showDeleteDialog?.displayName}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          if (!environment) {
            throw new Error('Project and Environment is required!.');
          }
          try {
            const { errors } = await api.deleteRouter({
              envName: environment,

              routerName: parseName(showDeleteDialog),
            });

            if (errors) {
              throw errors[0];
            }
            reloadPage();
            toast.success(`${titleCase(RESOURCE_NAME)} deleted successfully`);
            setShowDeleteDialog(null);
          } catch (err) {
            handleError(err);
          }
        }}
      />
      <HandleRouter
        {...{
          isUpdate: true,
          data: visible!,
          visible: !!visible,
          setVisible: () => setVisible(null),
        }}
      />
    </>
  );
};

export default RouterResourcesV2;
