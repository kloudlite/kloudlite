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
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
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
import { asyncPopupWindow, renderCloudProvider } from '~/console/utils/commons';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { Button, IconButton } from '~/components/atoms/button';
import Pulsable from '~/console/components/pulsable';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import Popup from '~/components/molecule/popup';
import CodeView from '~/console/components/code-view';
import Yup from '~/root/lib/server/helpers/yup';
import { PasswordInput } from '~/components/atoms/input';
import useForm from '~/root/lib/client/hooks/use-form';
import { Badge } from '~/components/atoms/badge';
import ListV2 from '~/console/components/listV2';
import ConsoleAvatar from '~/console/components/console-avatar';
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
  // const checkAwsAccess = async () => {
  //   const { data, errors } = await api.checkAwsAccess({
  //     cloudproviderName: item.metadata?.name || '',
  //   });
  //   if (errors) {
  //     throw errors[0];
  //   }
  //   return data;
  // };

  const [isLoading, setIsLoading] = useState(false);

  const { data, isLoading: il } = useCustomSwr(
    () => parseName(item) + isLoading,
    async () => {
      if (!parseName(item)) {
        throw new Error('Invalid cloud provider name');
      }
      return api.checkAwsAccess({
        cloudproviderName: parseName(item),
      });
    }
  );

  const { values, handleChange, errors, handleSubmit } = useForm({
    initialValues: {
      accessKey: '',
      secretKey: '',
    },
    validationSchema: Yup.object({
      accessKey: Yup.string().test(
        'provider',
        'access key is required',
        // @ts-ignores
        // eslint-disable-next-line react/no-this-in-sfc
        function (item) {
          return data?.result || item;
        }
      ),
      secretKey: Yup.string().test(
        'provider',
        'secret key is required',
        // eslint-disable-next-line func-names
        // @ts-ignore
        function (item) {
          return data?.result || item;
        }
      ),
    }),
    onSubmit: async (val) => {
      if (data?.result) {
        // navigate(
        //   `/onboarding/${parseName(account)}/${parseName(
        //     cloudProvider
        //   )}/new-cluster`
        // );
        toast.success('Provider validated successfully');

        onClose();
        return;
      }

      try {
        const { errors } = await api.updateProviderSecret({
          secret: {
            metadata: {
              name: parseName(item),
            },
            cloudProviderName: item.cloudProviderName,
            displayName: item.displayName,
            aws: {
              authMechanism: 'secret_keys',
              authSecretKeys: {
                accessKey: val.accessKey,
                secretKey: val.secretKey,
              },
            },
          },
        });

        if (errors) {
          throw errors[0];
        }

        setIsLoading((s) => !s);
      } catch (err) {
        handleError(err);
      }
    },
  });

  return (
    <Popup.Root onOpenChange={onClose} show={show}>
      <Popup.Header>Validate Aws Provider</Popup.Header>
      <Popup.Content>
        <form onSubmit={handleSubmit} className="flex flex-col gap-2xl">
          {/* <div className="flex gap-xl items-center"> */}
          {/*   <span>Account ID</span> */}
          {/*   <span>{item.aws?.awsAccountId}</span> */}
          {/* </div> */}
          {!data?.result && (
            <div className="flex flex-col gap-xl text-start">
              <CodeView copy data={url} />

              <span className="flex flex-wrap items-center gap-md">
                visit the link above and click on the button to validate your
                AWS account, or
                <Button
                  loading={il}
                  variant="primary-plain"
                  onClick={async () => {
                    setIsLoading(true);
                    try {
                      await asyncPopupWindow({ url });

                      setIsLoading((s) => !s);
                    } catch (err) {
                      handleError(err);
                    }

                    setIsLoading(false);
                  }}
                  content="click here"
                />
              </span>
            </div>
          )}

          {data?.result && (
            <div className="py-2xl">
              <Badge type="success" icon={<Check />}>
                Your Credential is valid
              </Badge>
            </div>
          )}

          {!data?.result && (
            <>
              <div className="">
                Once you have created the cloudformation stack, please enter the
                access key and secret key below to validate your cloud Provider,
                you can get the access key and secret key from the output of the
                cloudformation stack.
              </div>

              <PasswordInput
                name="accessKey"
                onChange={handleChange('accessKey')}
                error={!!errors.accessKey}
                message={errors.accessKey}
                value={values.accessKey}
                label="Access Key"
              />

              <PasswordInput
                name="secretKey"
                onChange={handleChange('secretKey')}
                error={!!errors.secretKey}
                message={errors.secretKey}
                value={values.secretKey}
                label="Secret Key"
              />

              <Button
                loading={il}
                variant="primary"
                content="Update"
                type="submit"
              />
            </>
          )}
        </form>
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
            {/* <span>{item.aws?.awsAccountId}</span> */}
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
          {/* <span>{item.aws?.awsAccountId}</span> */}
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

const ExtraButton = ({
  onDelete,
  onEdit,
}: {
  onDelete: () => void;
  onEdit: () => void;
}) => {
  const iconSize = 16;
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Edit',
          icon: <PencilSimple size={iconSize} />,
          type: 'item',
          onClick: onEdit,
          key: 'edit',
        },
        {
          label: 'Delete',
          icon: <Trash size={iconSize} />,
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

const GridView = ({ items = [], onDelete, onEdit }: IResource) => {
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
                  <ListTitle
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
                  <ListBody data={renderCloudProvider({ cloudprovider })} />
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

const ListView = ({ items = [], onDelete, onEdit }: IResource) => {
  return (
    <ListV2.Root
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
            render: () => 'Provider',
            name: 'provider',
            className: 'w-[180px]',
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
          const { name, id, cloudprovider, updateInfo } = parseItem(i);
          return {
            columns: {
              name: {
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              status: {
                render: () => (
                  <ListBody data={i.aws ? <AwsCheckBody item={i} /> : null} />
                ),
              },
              provider: {
                render: () => (
                  <ListBody data={renderCloudProvider({ cloudprovider })} />
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
                render: () => (
                  <ExtraButton
                    onDelete={() => onDelete(i)}
                    onEdit={() => onEdit(i)}
                  />
                ),
              },
            },
          };
        }),
      }}
    />
  );
};

const ProviderResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [showHandleProvider, setShowHandleProvider] = useState<BaseType | null>(
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
      setShowHandleProvider(item);
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
        {...{
          isUpdate: true,
          data: showHandleProvider!,
          setVisible: () => setShowHandleProvider(null),
          visible: !!showHandleProvider,
        }}
      />
    </>
  );
};

export default ProviderResourcesV2;
