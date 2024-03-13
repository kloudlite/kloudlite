import {
  GearSix,
  LinkBreak,
  Link as LinkIcon,
  Repeat,
} from '@jengaicons/react';
import { Link, useOutletContext, useParams } from '@remix-run/react';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListItem,
  ListSecondary,
  ListTitle,
  listClass,
  listFlex,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction, {
  IResourceExtraItem,
} from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IApps } from '~/console/server/gql/queries/app-queries';
import {
  ExtractNodeType,
  parseName,
  parseName as pn,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import { IEnvironmentContext } from '../_layout';

const RESOURCE_NAME = 'app';
type BaseType = ExtractNodeType<IApps>;

const parseItem = (item: ExtractNodeType<IApps>) => {
  return {
    name: item.displayName,
    id: pn(item),
    intercept: item.spec.intercept,
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

const ExtraButton = ({ onAction, item }: IExtraButton) => {
  const { account, project, environment } = useParams();
  const iconSize = 16;
  let options: IResourceExtraItem[] = [
    {
      label: 'Settings',
      icon: <GearSix size={iconSize} />,
      type: 'item',
      to: `/${account}/${project}/${environment}/app/${parseName(
        item
      )}/settings/general`,
      key: 'settings',
    },
  ];

  if (item.spec.intercept && item.spec.intercept.enabled) {
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

  options = [
    {
      label: 'Restart',
      icon: <LinkIcon size={iconSize} />,
      type: 'item',
      onClick: () => onAction({ action: 'restart', item }),
      key: 'restart',
    },
    ...options,
  ];

  return <ResourceExtraAction options={options} />;
};

interface IResource {
  items: BaseType[];
  onAction: OnAction;
}

const GridView = ({ items = [], onAction: _ }: IResource) => {
  const { account, project, environment } = useParams();

  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/${project}/${environment}/app/${id}`}
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
                            to: `/${account}/${project}/${environment}/app/${id}/settings/general`,
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
  const { account, project, environment } = useParams();
  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo, intercept } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            to={`/${account}/${project}/${environment}/app/${id}`}
            key={keyPrefix + name}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: listClass.title,
                render: () => <ListTitle title={name} subtitle={id} />,
              },
              // @ts-ignore
              ...[
                intercept && !!intercept.enabled
                  ? {
                      key: generateKey(keyPrefix, `${name + id}intercept`),
                      className: listClass.title,
                      render: () => (
                        <ListSecondary
                          title="intercepted toDevice"
                          subtitle={intercept?.toDevice}
                        />
                      ),
                    }
                  : [],
              ],
              listFlex({ key: 'flex-1' }),
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
              {
                key: generateKey(keyPrefix, 'action'),
                render: () => <ExtraButton onAction={onAction} item={item} />,
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const AppsResources = ({ items = [] }: Omit<IResource, 'onAction'>) => {
  const api = useConsoleApi();
  const { environment, project, account } = useParams();
  const { devicesForUser } = useOutletContext<IEnvironmentContext>();
  const reload = useReload();

  useWatchReload(
    items.map((i) => {
      return `account:${account}.project:${project}.environment:${environment}.app:${parseName(
        i
      )}`;
    })
  );

  // useWatchItems(items, (item) => ({
  //   account,
  //   project,
  //   environment,
  //   app: pn(item),
  // }));

  const interceptApp = async (item: BaseType, intercept: boolean) => {
    if (!environment || !project) {
      throw new Error('Environment is required!.');
    }
    if (devicesForUser && devicesForUser.length > 0) {
      const device = devicesForUser[0];
      try {
        const { errors } = await api.interceptApp({
          appname: pn(item),
          deviceName: pn(device),
          envName: environment,
          intercept,
          projectName: project,
        });

        if (errors) {
          throw errors[0];
        }
        toast.success('App intercepted successfully');
        reload();
      } catch (error) {
        handleError(error);
      }
    }
  };

  const restartApp = async (item: BaseType) => {
    if (!environment || !project) {
      throw new Error('Environment is required!.');
    }

    try {
      const { errors } = await api.restartApp({
        appName: pn(item),
        envName: environment,
        projectName: project,
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
          interceptApp(item, true);
          break;
        case 'restart':
          restartApp(item);
          break;
        case 'remove_intercept':
          interceptApp(item, false);
          break;
        default:
      }
    },
  };
  return (
    <ListGridView
      listView={<ListView {...props} />}
      gridView={<GridView {...props} />}
    />
  );
};

export default AppsResources;
