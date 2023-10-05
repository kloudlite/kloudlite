import { DotsThreeVerticalFill } from '@jengaicons/react';
import { Link } from '@remix-run/react';
import { IconButton } from '~/components/atoms/button';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithAvatar,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import {
  IManagedServiceTemplates,
  IManagedServices,
} from '~/console/server/gql/queries/managed-service-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { getManagedTemplate } from '~/console/utils/commons';

const RESOURCE_NAME = 'backend service';

const parseItem = (
  item: ExtractNodeType<IManagedServices>,
  templates: IManagedServiceTemplates
) => {
  const template = getManagedTemplate({
    templates,
    kind: item.spec.msvcKind.kind || '',
    apiVersion: item.spec.msvcKind.apiVersion,
  });
  return {
    name: item?.displayName,
    id: parseName(item),
    type: item?.kind,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
    logo: template?.logoUrl,
  };
};

const GridView = ({
  items = [],
  templates = [],
}: {
  items: ExtractNodeType<IManagedServices>[];
  templates: ExtractNodeType<IManagedServiceTemplates>;
}) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, type, logo, updateInfo } = parseItem(item, templates);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            to={`../backing-service/${id}`}
            key={id}
            rows={[
              {
                key: generateKey(keyPrefix, name),
                render: () => (
                  <ListTitleWithAvatar
                    title={name}
                    action={
                      <IconButton
                        icon={<DotsThreeVerticalFill />}
                        variant="plain"
                        onClick={(e) => e.stopPropagation()}
                      />
                    }
                    avatar={
                      <img src={logo} alt={name} className="w-4xl h-4xl" />
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, type),
                render: () => (
                  <div className="flex flex-col gap-md">
                    <ListBody data={type} />
                  </div>
                ),
              },
              {
                key: generateKey(keyPrefix, 'author'),
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

const ListView = ({
  items = [],
  templates = [],
}: {
  items: ExtractNodeType<IManagedServices>[];
  templates: ExtractNodeType<IManagedServiceTemplates>;
}) => {
  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, type, logo, updateInfo } = parseItem(item, templates);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;

        return (
          <List.Row
            to={`../backing-service/${id}`}
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name),
                className: 'flex-1',
                render: () => (
                  <ListTitleWithAvatar
                    title={name}
                    avatar={
                      <img src={logo} alt={name} className="w-4xl h-4xl" />
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, type),
                className: 'w-[140px] text-start',
                render: () => <ListBody data={type} />,
              },
              {
                key: generateKey(keyPrefix, 'author'),
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
                  <IconButton
                    icon={<DotsThreeVerticalFill />}
                    variant="plain"
                    onClick={(e) => e.stopPropagation()}
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

const BackendServicesResources = ({
  items = [],
  templates = [],
}: {
  items: ExtractNodeType<IManagedServices>[];
  templates: ExtractNodeType<IManagedServiceTemplates>;
}) => {
  return (
    <ListGridView
      listView={<ListView items={items} templates={templates} />}
      gridView={<GridView items={items} templates={templates} />}
    />
  );
};

export default BackendServicesResources;
