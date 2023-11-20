import {
  ArrowCounterClockwise,
  PencilSimple,
  Trash,
  Check,
} from '@jengaicons/react';
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
import { IShowDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IProviderSecrets } from '~/console/server/gql/queries/provider-secret-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { DIALOG_TYPE, asyncPopupWindow } from '~/console/utils/commons';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { Button, IconButton } from '~/components/atoms/button';
import Pulsable from '~/console/components/pulsable';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import Popup from '~/components/molecule/popup';
import CodeView from '~/console/components/code-view';
import HandleProvider from './handle-provider';

const RESOURCE_NAME = 'cloud provider';
type BaseType = ExtractNodeType<IProviderSecrets>;

const AwsValidationPopup = ({
  show,
  item,
  onClose,
  url,
}: {
  show: boolean;
  item: BaseType;
  onClose: () => void;
  url: string;
}) => {
  const api = useConsoleApi();
  const checkAwsAccess = async () => {
    const { data, errors } = await api.checkAwsAccess({
      cloudproviderName: item.metadata?.name || '',
    });
    if (errors) {
      throw errors[0];
    }
    return data;
  };

  const [isLoading, setIsLoading] = useState(false);

  return (
    <Popup.Root onOpenChange={onClose} show={show}>
      <Popup.Header>Validate Aws Provider</Popup.Header>
      <Popup.Content>
        <div className="flex flex-col gap-2xl">
          <div className="flex gap-xl items-center">
            <span>Account ID</span>
            <span>{item.aws?.awsAccountId}</span>
          </div>
          <div className="flex flex-col gap-xl text-start">
            <CodeView copy data={url} />

            <span className="flex flex-wrap items-center gap-md">
              visit the link above and click on the button to validate your AWS
              account, or
              <Button
                loading={isLoading}
                variant="primary-plain"
                onClick={async () => {
                  setIsLoading(true);
                  try {
                    await asyncPopupWindow({ url });

                    const res = await checkAwsAccess();

                    if (res.result) {
                      toast.success('Aws account validated successfully');
                    } else {
                      toast.error('Aws account validation failed');
                    }
                  } catch (err) {
                    handleError(err);
                  }

                  setIsLoading(false);
                }}
                content="click here"
              />
            </span>
          </div>
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Button variant="primary-outline" content="close" onClick={onClose} />
      </Popup.Footer>
    </Popup.Root>
  );
};

const AwsCheckBodyWithValidation = ({ item }: { item: BaseType }) => {
  const api = useConsoleApi();

  const [show, setShow] = useState(false);

  const { data, isLoading } = useCustomSwr(
    item.metadata?.name || null,
    async () => {
      if (!item.metadata?.name) {
        throw new Error('Invalid cloud provider name');
      }
      return api.checkAwsAccess({
        cloudproviderName: item.metadata.name,
      });
    }
  );

  return (
    <Pulsable isLoading={isLoading}>
      <div className="pulsable">
        {data?.result ? (
          <div className="flex gap-xl items-center pulsable">
            <span>{item.aws?.awsAccountId}</span>
            <Button
              size="sm"
              variant="primary-outline"
              content={<Check size={16} />}
            />
          </div>
        ) : (
          <div>
            <Button
              onClick={() => setShow((s) => !s)}
              variant="critical-outline"
              size="sm"
              content="Invalid"
            />
            <AwsValidationPopup
              url={data?.installationUrl || ''}
              show={show}
              onClose={() => {
                setShow(false);
              }}
              item={item}
            />
          </div>
        )}
      </div>
    </Pulsable>
  );
};

const AwsCheckBody = ({ item }: { item: BaseType }) => {
  const [show, setShow] = useState(false);

  return (
    <div>
      {show ? (
        <AwsCheckBodyWithValidation item={item} />
      ) : (
        <div className="flex gap-xl items-center pulsable">
          <span>{item.aws?.awsAccountId}</span>
          <IconButton
            onClick={() => {
              setShow(true);
            }}
            size="sm"
            variant="outline"
            icon={<ArrowCounterClockwise size={16} />}
          />
        </div>
      )}
    </div>
  );
};

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: parseName(item),
    cloudprovider: item.cloudProviderName,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

interface IExtraButton {
  onDelete: () => void;
  onEdit: () => void;
}
const ExtraButton = ({ onDelete, onEdit }: IExtraButton) => {
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Edit',
          icon: <PencilSimple size={16} />,
          type: 'item',
          onClick: onEdit,
          key: 'edit',
        },
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
  items: BaseType[];
  onDelete: (item: BaseType) => void;
  onEdit: (item: BaseType) => void;
}

const GridView = ({
  items = [],
  onDelete = (_) => _,
  onEdit = (_) => _,
}: IResource) => {
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
                    action={
                      <ExtraButton
                        onDelete={() => onDelete(item)}
                        onEdit={() => onEdit(item)}
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
  items = [],
  onDelete = (_) => _,
  onEdit = (_) => _,
}: IResource) => {
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
                key: generateKey(keyPrefix, name + id + cloudprovider),
                className: 'text-start',
                render: () => (
                  <ListBody
                    data={
                      item.aws?.awsAccountId ? (
                        <AwsCheckBody item={item} />
                      ) : null
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, cloudprovider),
                className: 'w-[120px] text-start',
                render: () => <ListBody data={cloudprovider} />,
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
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
                  <ExtraButton
                    onDelete={() => onDelete(item)}
                    onEdit={() => onEdit(item)}
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

const ProviderResources = ({ items = [] }: { items: BaseType[] }) => {
  const [showHandleProvider, setShowHandleProvider] =
    useState<IShowDialog<BaseType | null>>(null);
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    onDelete: (item) => {
      setShowDeleteDialog(item);
    },

    onEdit: (item) => {
      setShowHandleProvider({ type: DIALOG_TYPE.EDIT, data: item });
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
      <HandleProvider
        show={showHandleProvider}
        setShow={setShowHandleProvider}
      />
    </>
  );
};

export default ProviderResources;
