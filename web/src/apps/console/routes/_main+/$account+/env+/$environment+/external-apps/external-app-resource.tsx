import { GearSix, LinkBreak, Repeat } from '~/console/components/icons';
import { Link, useOutletContext, useParams } from '@remix-run/react';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction, {
  IResourceExtraItem,
} from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import {
  ExtractNodeType,
  parseName,
  parseName as pn,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { handleError } from '~/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { useReload } from '~/lib/client/helpers/reloader';
import { SyncStatusV2 } from '~/console/components/sync-status';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import ListV2 from '~/console/components/listV2';
import { useState } from 'react';
import { Badge } from '~/components/atoms/badge';
import { CopyContentToClipboard } from '~/console/components/common-console-components';
import Tooltip from '~/components/atoms/tooltip';
import { NN } from '~/root/lib/types/common';
// import HandleIntercept from './handle-intercept';
import { IExternalApps } from '~/console/server/gql/queries/external-app-queries';
import { IEnvironmentContext } from '../_layout';
import HandleExternalAppIntercept from './handle-external-app-intercept';

const RESOURCE_NAME = 'app';
type BaseType = ExtractNodeType<IExternalApps>;

const parseItem = (item: ExtractNodeType<IExternalApps>) => {
  return {
    name: item.displayName,
    id: pn(item),
    intercept: item.spec?.intercept,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

type OnAction = ({
  action,
  item,
}: {
  action: 'delete' | 'edit' | 'intercept' | 'remove_intercept' | 'restart';
  item: BaseType;
}) => void;

type IExtraButton = {
  onAction: OnAction;
  item: BaseType;
};

const InterceptPortView = ({
  ports = [],
  devName = '',
}: {
  ports: NN<
    NN<ExtractNodeType<IExternalApps>['spec']>['intercept']
  >['portMappings'];
  devName: string;
}) => {
  return (
    <div className="flex flex-row items-center gap-md">
      <Tooltip.Root
        align="start"
        side="top"
        className="!max-w-fit "
        content={
          <div>
            <span className="bodyMd-medium text-text-soft">
              {' '}
              intercepted to{' '}
              <span className="bodyMd-medium text-text-strong">{devName}</span>
            </span>
            <div className="flex flex-row gap-md py-md">
              {ports?.map((d) => {
                return (
                  <Badge className="shrink-0" key={d.appPort}>
                    <div>
                      {d.appPort} → {d.devicePort}
                    </div>
                  </Badge>
                );
              })}
            </div>
          </div>
        }
      >
        <div className="bodyMd-medium text-text-strong w-fit truncate">
          {ports?.length === 1 ? (
            <span>{ports.length} port</span>
          ) : (
            <span>{ports.length} ports</span>
          )}
          <span className="text-text-soft">
            {' '}
            intercepted to{' '}
            <span className="bodyMd-medium text-text-strong truncate">
              {devName}
            </span>
          </span>
        </div>
      </Tooltip.Root>
    </div>
  );
};

const ExtraButton = ({ onAction, item }: IExtraButton) => {
  const { account, environment } = useParams();
  const iconSize = 16;
  let options: IResourceExtraItem[] = [
    {
      label: 'Settings',
      icon: <GearSix size={iconSize} />,
      type: 'item',
      to: `/${account}/env/${environment}/external-app/${parseName(
        item
      )}/settings/general`,
      key: 'settings',
    },
  ];

  if (item.spec?.intercept && item.spec.intercept.enabled) {
    options = [
      {
        label: 'Remove intercept',
        icon: <LinkBreak size={iconSize} />,
        type: 'item',
        onClick: () => onAction({ action: 'remove_intercept', item }),
        key: 'remove-intercept',
      },
      ...options,
    ];
  } else {
    options = [
      {
        label: 'Intercept',
        icon: <Repeat size={iconSize} />,
        type: 'item',
        onClick: () => onAction({ action: 'intercept', item }),
        key: 'intercept',
      },
      ...options,
    ];
  }

  // options = [
  //   {
  //     label: 'Restart',
  //     icon: <LinkIcon size={iconSize} />,
  //     type: 'item',
  //     onClick: () => onAction({ action: 'restart', item }),
  //     key: 'restart',
  //   },
  //   ...options,
  // ];

  return <ResourceExtraAction options={options} />;
};

interface IResource {
  items: BaseType[];
  onAction: OnAction;
}

const AppServiceView = ({ service }: { service: string }) => {
  return (
    <CopyContentToClipboard
      content={service}
      toastMessage="App service url copied successfully."
    />
  );
};

const GridView = ({ items = [], onAction: _ }: IResource) => {
  const { account, environment } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/env/${environment}/app/${id}`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={
                      <ResourceExtraAction
                        options={[
                          {
                            key: 'apps-resource-extra-action-1',
                            to: `/${account}/env/${environment}/app/${id}/settings/general`,
                            icon: <GearSix size={16} />,
                            label: 'Settings',
                            type: 'item',
                          },
                        ]}
                      />
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

const ListView = ({ items = [], onAction }: IResource) => {
  const { environment, cluster, account } =
    useOutletContext<IEnvironmentContext>();
  return (
    <ListV2.Root
      linkComponent={Link}
      data={{
        headers: [
          {
            render: () => 'Name',
            name: 'name',
            className: 'w-[180px]',
          },
          {
            render: () => '',
            name: 'intercept',
            className: 'w-[250px] truncate',
          },
          {
            render: () => '',
            name: 'flex-pre',
            className: 'flex-1',
          },
          {
            render: () => 'Service',
            name: 'service',
            className: 'w-[240px] flex flex-1',
          },
          {
            render: () => '',
            name: 'flex-post',
            className: 'flex-1',
          },
          {
            render: () => 'Status',
            name: 'status',
            className: 'w-[180px] ',
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
                render: () => <ListTitle title={name} subtitle={id} />,
              },
              intercept: {
                render: () =>
                  i.spec?.intercept?.enabled ? (
                    <div>
                      <InterceptPortView
                        ports={i.spec?.intercept.portMappings || []}
                        devName={i.spec?.intercept.toDevice || ''}
                      />
                    </div>
                  ) : null,
              },
              service: {
                render: () => (
                  <AppServiceView
                    service={
                      environment?.spec?.targetNamespace
                        ? `${parseName(i)}.${
                            environment?.spec?.targetNamespace
                          }.svc.${parseName(cluster)}.local`
                        : `${parseName(i)}.${parseName(
                            environment
                          )}.svc.${parseName(cluster)}.local`
                    }
                  />
                ),
              },
              status: {
                render: () => (
                  <div className="inline-block">
                    <SyncStatusV2 item={i} />
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
                render: () => <ExtraButton onAction={onAction} item={i} />,
              },
            },
            to: `/${parseName(account)}/env/${parseName(
              environment
            )}/external-app/${id}`,
          };
        }),
      }}
    />
  );
};

const ExternalNameResource = ({ items = [] }: Omit<IResource, 'onAction'>) => {
  const api = useConsoleApi();
  const { environment, account } = useOutletContext<IEnvironmentContext>();
  const reload = useReload();

  const [visible, setVisible] = useState(false);
  const [mi, setItem] = useState<ExtractNodeType<IExternalApps>>();

  useWatchReload(
    items.map((i) => {
      return `account:${parseName(account)}.environment:${parseName(
        environment
      )}.app:${parseName(i)}`;
    })
  );

  const interceptExternalApp = async (item: BaseType, intercept: boolean) => {
    if (intercept) {
      setItem(item);
      setVisible(true);
      return;
    }

    try {
      const { errors } = await api.interceptExternalApp({
        externalAppName: pn(item),
        deviceName: item.spec?.intercept?.toDevice || '',
        envName: pn(environment),
        intercept,
      });

      if (errors) {
        throw errors[0];
      }
      toast.success('external app intercept removed successfully');
      reload();
    } catch (error) {
      handleError(error);
    }
  };

  const restartApp = async (item: BaseType) => {
    if (!environment) {
      throw new Error('Environment is required!.');
    }

    try {
      const { errors } = await api.restartApp({
        appName: pn(item),
        envName: pn(environment),
      });

      if (errors) {
        throw errors[0];
      }
      toast.success('App restarted successfully');
      // reload();
    } catch (error) {
      handleError(error);
    }
  };

  const props: IResource = {
    items,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'intercept':
          interceptExternalApp(item, true);
          break;
        case 'restart':
          restartApp(item);
          break;
        case 'remove_intercept':
          interceptExternalApp(item, false);
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
      <HandleExternalAppIntercept
        {...{
          visible,
          setVisible,
          app: mi,
        }}
      />
    </>
  );
};

export default ExternalNameResource;
