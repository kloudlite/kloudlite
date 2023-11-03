import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import * as Chips from '~/components/atoms/chips';
import { PasswordInput, TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select-primitive';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { IdSelector } from '~/console/components/id-selector';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IProviderSecrets } from '~/console/server/gql/queries/provider-secret-queries';
import {
  ExtractNodeType,
  parseName,
  validateCloudProvider,
} from '~/console/server/r-utils/common';
import { DIALOG_TYPE } from '~/console/utils/commons';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { InfoLabel } from '../_.$account.$cluster.$project.$scope.$workspace.new-app/util';

type IHMap = {
  [key: string]: any;
};

const AwsForm = ({
  errors,
  values,
  handleChange,
}: {
  errors: IHMap;
  values: IHMap;
  handleChange: (name: string) => (e: any) => void;
}) => {
  const [withAccId, setWithAccId] = useState(false);

  return (
    <>
      {withAccId ? (
        <div className="flex-1">
          <TextInput
            name="awsAccountId"
            onChange={handleChange('awsAccountId')}
            error={!!errors.awsAccountId}
            message={errors.awsAccountId}
            value={values.awsAccountId}
            label="Account ID"
          />
        </div>
      ) : (
        <>
          <PasswordInput
            name="accessKey"
            onChange={handleChange('accessKey')}
            error={!!errors.accessKey}
            message={errors.accessKey}
            value={values.accessKey}
            label={
              <InfoLabel
                info={
                  <div>
                    <p>
                      Provide access key and secret key to access your AWS
                      account. <br />
                      We need these creds with following permissions: <br />
                    </p>
                    <ul className="px-md">
                      <li>ec2</li>
                      <li>s3</li>
                      <li>spotFleetTaggingRole</li>
                    </ul>
                  </div>
                }
                label="Access Key ID"
              />
            }
          />
          <PasswordInput
            name="accessSecret"
            label="Access Key Secret"
            onChange={handleChange('accessSecret')}
            error={!!errors.accessSecret}
            message={errors.accessSecret}
            value={values.accessSecret}
          />
        </>
      )}

      <div className="flex">
        <Button
          onClick={() => {
            return setWithAccId((s) => !s);
          }}
          variant="primary-plain"
          content={
            withAccId ? 'Use access creds instead' : 'Use account id instead'
          }
        />
      </div>
    </>
  );
};

const HandleProvider = ({
  show,
  setShow,
}: IDialog<ExtractNodeType<IProviderSecrets> | null, null>) => {
  const api = useConsoleApi();
  const reloadPage = useReload();

  const [validationSchema, setValidationSchema] = useState(
    Yup.object({
      displayName: Yup.string().required(),
      name: Yup.string().required(),
      provider: Yup.string().required(),
      // accessKey: Yup.string().required(),
      // accessSecret: Yup.string().required(),
    })
  );

  const {
    values,
    errors,
    handleSubmit,
    handleChange,
    isLoading,
    resetValues,
    setValues,
  } = useForm({
    initialValues: {
      displayName: '',
      name: '',
      provider: 'aws',
      accessKey: '',
      accessSecret: '',
      awsAccountId: '',
    },
    validationSchema,

    onSubmit: async (val) => {
      // const validateAccountIdAndPerform = async <T,>(
      //   fn: () => T
      // ): Promise<T> => {
      //   const { data, errors } = await api.checkAwsAccess({
      //     accountId: val.accountId,
      //   });
      //
      //   if (errors) {
      //     throw errors[0];
      //   }
      //
      //   if (!data.result) {
      //     await asyncPopupWindow({
      //       url: data.installationUrl || '',
      //     });
      //
      //     const { data: d2 } = await api.checkAwsAccess({
      //       accountId: val.accountId,
      //     });
      //
      //     if (!d2.result) {
      //       throw new Error('invalid account id');
      //     }
      //
      //     return fn();
      //   }
      //
      //   return fn();
      // };

      const addProvider = async () => {
        switch (val.provider) {
          case 'aws':
            if (val.awsAccountId) {
              // return validateAccountIdAndPerform(async () => {
              // });

              return api.createProviderSecret({
                secret: {
                  displayName: val.displayName,
                  metadata: {
                    name: val.name,
                  },
                  stringData: {
                    awsAccountId: val.awsAccountId,
                  },
                  cloudProviderName: validateCloudProvider(val.provider),
                },
              });
            }

            return api.createProviderSecret({
              secret: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },
                stringData: {
                  accessKey: val.accessKey,
                  accessSecret: val.accessSecret,
                },
                cloudProviderName: validateCloudProvider(val.provider),
              },
            });

          default:
            throw new Error('invalid provider');
        }
      };

      const updateProvider = async () => {
        if (!show?.data) {
          throw new Error("data can't be null");
        }

        switch (val.provider) {
          case 'aws':
            if (val.awsAccountId) {
              // return validateAccountIdAndPerform(async () => {
              //   if (!show?.data) {
              //     throw new Error("data can't be null");
              //   }
              //
              // });

              return api.updateProviderSecret({
                secret: {
                  cloudProviderName: show.data.cloudProviderName,
                  displayName: val.displayName,
                  metadata: {
                    name: parseName(show.data, true),
                  },
                  stringData: {
                    awsAccountId: val.awsAccountId,
                  },
                },
              });
            }

            return api.updateProviderSecret({
              secret: {
                cloudProviderName: show.data.cloudProviderName,
                displayName: val.displayName,
                metadata: {
                  name: parseName(show.data, true),
                },
                stringData: {
                  secretKey: val.accessSecret,
                  accessKey: val.accessKey,
                },
              },
            });
          default:
            throw new Error('invalid provider');
        }
      };

      try {
        if (show?.type === DIALOG_TYPE.ADD) {
          const { errors: e } = await addProvider();
          if (e) {
            throw e[0];
          }
          toast.success('provider secret created successfully');
        } else if (show?.data) {
          const { errors: e } = await updateProvider();
          if (e) {
            throw e[0];
          }
        }
        reloadPage();
        setShow(null);
        resetValues();
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    if (show && show.data && show.type === DIALOG_TYPE.EDIT) {
      setValues((v) => ({
        ...v,
        displayName: show.data?.displayName || '',
        accessSecret: '',
        accessKey: '',
      }));
      setValidationSchema(
        // @ts-ignore
        Yup.object({
          displayName: Yup.string().trim().required(),
          // accessSecret: Yup.string().trim().required(),
          // accessKey: Yup.string().trim().required(),
          provider: Yup.string().required(),
        })
      );
    }
  }, [show]);

  return (
    <Popup.Root
      show={!!show}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }
        setShow(e);
      }}
    >
      <Popup.Header>
        {show?.type === DIALOG_TYPE.ADD
          ? 'Add new cloud provider'
          : 'Edit cloud provider'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            {show?.type === DIALOG_TYPE.EDIT && (
              <Chips.Chip
                {...{
                  item: { id: parseName(show.data) },
                  label: parseName(show.data),
                  prefix: 'Id:',
                  disabled: true,
                  type: 'BASIC',
                }}
              />
            )}

            <TextInput
              label="Name"
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
              value={values.displayName}
              name="provider-secret-name"
            />
            {show?.type === DIALOG_TYPE.ADD && (
              <IdSelector
                name={values.displayName}
                resType="providersecret"
                onChange={(id) => {
                  handleChange('name')({ target: { value: id } });
                }}
              />
            )}
            {show?.type === DIALOG_TYPE.ADD && (
              <Select.Root
                error={!!errors.provider}
                message={errors.provider}
                value={values.provider}
                label="Provider"
                onChange={handleChange('provider')}
              >
                <Select.Option value="aws">Amazon Web Services</Select.Option>
              </Select.Root>
            )}

            {values.provider === 'aws' && (
              <AwsForm
                {...{
                  values,
                  errors,
                  handleChange,
                }}
              />
            )}
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button content="Cancel" variant="basic" closable />
          <Popup.Button
            loading={isLoading}
            type="submit"
            content={show?.type === DIALOG_TYPE.ADD ? 'Add' : 'Update'}
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

export default HandleProvider;
