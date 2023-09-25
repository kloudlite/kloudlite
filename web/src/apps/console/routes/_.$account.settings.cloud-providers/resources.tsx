import { DotsThreeVerticalFill, Info } from '@jengaicons/react';
import { Badge } from '~/components/atoms/badge';
import { IconButton } from '~/components/atoms/button';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithSubtitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import { IProviderSecrets } from '~/console/server/gql/queries/provider-secret-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';

const parseItem = (item: ExtractNodeType<IProviderSecrets>) => {
  return {
    name: item.displayName,
    id: parseName(item),
    cloudprovider: item.cloudProviderName,
    path: `/projects/${parseName(item)}`,
    running: item.status?.isReady,
    updateInfo: {
      author: titleCase(
        `${parseUpdateOrCreatedBy(item)} updated the cloud provider`
      ),
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const GridView = ({
  items = [],
}: {
  items: ExtractNodeType<IProviderSecrets>[];
}) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
        const { name, id, running, cloudprovider, updateInfo } =
          parseItem(item);
        const keyPrefix = `cloudprovider-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
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
                key: generateKey(keyPrefix, cloudprovider),
                render: () => (
                  <div className="flex flex-col gap-2xl">
                    <ListBody data={cloudprovider} />
                    <ListBody
                      data={
                        <Badge
                          icon={<Info />}
                          type={running ? 'neutral' : 'critical'}
                        >
                          {running ? 'Running' : 'Stopped'}
                        </Badge>
                      }
                    />
                  </div>
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

const ListView = ({
  items,
}: {
  items: ExtractNodeType<IProviderSecrets>[];
}) => {
  return (
    <List.Root>
      {items.map((item, index) => {
        const { name, id, running, cloudprovider, updateInfo } =
          parseItem(item);
        const keyPrefix = `cloudprovider-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'flex-1',
                render: () => (
                  <ListTitleWithSubtitle title={name} subtitle={id} />
                ),
              },
              {
                key: generateKey(keyPrefix, cloudprovider),
                className: 'w-[120px] text-start',
                render: () => <ListBody data={cloudprovider} />,
              },
              {
                key: generateKey(keyPrefix, 'status'),
                className: 'w-[120px]',
                render: () => (
                  <ListBody
                    data={
                      <Badge
                        icon={<Info />}
                        type={running ? 'neutral' : 'critical'}
                      >
                        {running ? 'Running' : 'Stopped'}
                      </Badge>
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

const Resources = ({
  items = [],
}: {
  items: ExtractNodeType<IProviderSecrets>[];
}) => {
  return (
    <ListGridView
      listView={<ListView items={items} />}
      gridView={<GridView items={items} />}
    />
  );
};

export default Resources;
