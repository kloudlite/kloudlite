import { Trash } from '@jengaicons/react';
import { useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithSubtitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IProviderSecrets } from '~/console/server/gql/queries/provider-secret-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';

const RESOURCE_NAME = 'cloud provider';

const parseItem = (item: ExtractNodeType<IProviderSecrets>) => {
  return {
    name: item.displayName,
    id: parseName(item),
    cloudprovider: item.cloudProviderName,
    path: `/projects/${parseName(item)}`,
    updateInfo: {
      author: titleCase(
        `${parseUpdateOrCreatedBy(item)} updated the ${RESOURCE_NAME}`
      ),
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

interface IExtraButton {
  onDelete: () => void;
}
const ExtraButton = ({ onDelete }: IExtraButton) => {
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Delete',
          icon: <Trash size={16} />,
          type: 'item',
          onClick: onDelete,
          key: 'delete',
          className: '!text-text-critical',
        },
      ]}
    />
  );
};

interface IResource {
  items: ExtractNodeType<IProviderSecrets>[];
  onDelete: (item: ExtractNodeType<IProviderSecrets>) => void;
}

const GridView = ({ items = [], onDelete = (_) => _ }: IResource) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
        const { name, id, cloudprovider, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
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
                    action={<ExtraButton onDelete={() => onDelete(item)} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, cloudprovider),
                render: () => (
                  <div className="flex flex-col gap-2xl">
                    <ListBody data={cloudprovider} />
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

const ListView = ({ items, onDelete = (_) => _ }: IResource) => {
  return (
    <List.Root>
      {items.map((item, index) => {
        const { name, id, cloudprovider, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
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
                render: () => <ExtraButton onDelete={() => onDelete(item)} />,
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
  const [showDeleteDialog, setShowDeleteDialog] =
    useState<ExtractNodeType<IProviderSecrets> | null>(null);

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    onDelete: (item) => {
      setShowDeleteDialog(item);
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
          try {
            const { errors } = await api.deleteProviderSecret({
              secretName: parseName(showDeleteDialog),
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
    </>
  );
};

export default Resources;
