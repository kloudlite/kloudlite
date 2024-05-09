import { GearSix } from '~/console/components/icons';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { IMSvTemplates } from '~/console/server/gql/queries/managed-templates-queries';
import { getManagedTemplate } from '~/console/utils/commons';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { Link, useOutletContext, useParams } from '@remix-run/react';
import { SyncStatusV2 } from '~/console/components/sync-status';
import ListV2 from '~/console/components/listV2';
import { IClusterMSvs } from '~/console/server/gql/queries/cluster-managed-services-queries';
import { IAccountContext } from '../_layout';

const RESOURCE_NAME = 'managed service';
type BaseType = ExtractNodeType<IClusterMSvs>;

const parseItem = (item: BaseType, templates: IMSvTemplates) => {
  const template = getManagedTemplate({
    templates,
    kind: item.spec?.msvcSpec?.serviceTemplate.kind || '',
    apiVersion: item.spec?.msvcSpec?.serviceTemplate.apiVersion || '',
  });
  return {
    name: item?.displayName,
    id: parseName(item),
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
    logo: template?.logoUrl,
  };
};

const ExtraButton = ({ managedService }: { managedService: BaseType }) => {
  const { account } = useParams();
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Settings',
          icon: <GearSix size={16} />,
          type: 'item',

          to: `/${account}/msvc/${parseName(managedService)}/settings`,
          key: 'settings',
        },
      ]}
    />
  );
};

const GridView = ({
  items,
  templates,
}: {
  items: BaseType[];
  templates: IMSvTemplates;
}) => {
  const { account, project } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, logo, updateInfo } = parseItem(item, templates);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/${project}/msvc/${id}/logs-n-metrics`}
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
                            key: 'managed-services-resource-extra-action-1',
                            to: `/${account}/${project}/msvc/${id}/logs-n-metrics`,
                            icon: <GearSix size={16} />,
                            label: 'logs & metrics',
                            type: 'item',
                          },
                        ]}
                      />
                    }
                    // action={<ExtraButton onAction={onAction} item={item} />}
                    avatar={
                      <img src={logo} alt={name} className="w-4xl h-4xl" />
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'author'),
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

const ListView = ({
  items,
  templates,
}: {
  items: BaseType[];
  templates: IMSvTemplates;
}) => {
  const { account } = useOutletContext<IAccountContext>();
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
            className: 'w-[180px]',
          },
          {
            render: () => 'Status',
            name: 'status',
            className: 'flex-1 min-w-[30px] flex items-center justify-center',
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
          const { name, id, logo, updateInfo } = parseItem(i, templates);
          return {
            columns: {
              name: {
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    avatar={
                      <div className="pulsable pulsable-circle aspect-square">
                        <img src={logo} alt={name} className="w-4xl h-4xl" />
                      </div>
                    }
                  />
                ),
              },
              status: {
                render: () => <SyncStatusV2 item={i} />,
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
                render: () => <ExtraButton managedService={i} />,
              },
            },
            to: `/${parseName(account)}/msvc/${id}/logs-n-metrics`,
          };
        }),
      }}
    />
  );
};

const BackendServicesResourcesV2 = ({
  items = [],
  templates = [],
}: {
  items: BaseType[];
  templates: IMSvTemplates;
}) => {
  return (
    <ListGridView
      listView={<ListView items={items} templates={templates} />}
      gridView={<GridView items={items} templates={templates} />}
    />
  );
};

export default BackendServicesResourcesV2;
