/* eslint-disable react/destructuring-assignment */
import { ReactNode } from 'react';
import * as Chips from '~/components/atoms/chips';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { IdSelector } from '~/console/components/id-selector';
import { IDialogBase } from '~/console/components/types.d';
import { AwsForm } from '~/console/page-components/cloud-provider';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IProviderSecrets } from '~/console/server/gql/queries/provider-secret-queries';
import {
  ExtractNodeType,
  parseName,
  validateCloudProvider,
} from '~/console/server/r-utils/common';
import { providerIcons } from '~/console/utils/commons';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

type IDialog = IDialogBase<ExtractNodeType<IProviderSecrets>>;

const valueRender = ({
  label,
  labelValueIcon,
}: {
  label: string;
  labelValueIcon: ReactNode;
}) => {
  return (
    <div className="flex flex-row gap-xl items-center bodyMd text-text-default">
      <span>{labelValueIcon}</span>
      <span>{label}</span>
    </div>
  );
};

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();
  const iconSize = 16;

  const providers = [
    {
      label: 'Amazon Web Services',
      value: 'aws',
      labelValueIcon: providerIcons(iconSize).aws,
      render: () => (
        <div className="flex flex-row gap-lg items-center">
          <div>{providerIcons(iconSize).aws}</div>
          <div>Add Github Account</div>
        </div>
      ),
    },
  ];

  const { values, errors, handleSubmit, handleChange, isLoading, resetValues } =
    useForm({
      initialValues: isUpdate
        ? {
            displayName: props.data.displayName,
            name: parseName(props.data),
            provider: providers.find(
              (p) => p.value === props.data.cloudProviderName
            ),
            accessKey: '',
            accessSecret: '',
            awsAccountId: '',
          }
        : {
            displayName: '',
            name: '',
            provider: providers[0],
            accessKey: '',
            accessSecret: '',
            awsAccountId: '',
          },
      validationSchema: isUpdate
        ? Yup.object({
            displayName: Yup.string().required(),
            name: Yup.string().required(),
          })
        : Yup.object({
            displayName: Yup.string().required(),
            name: Yup.string().required(),
            provider: Yup.object({
              label: Yup.string().required(),
              value: Yup.string().required(),
            }).required(),
          }),

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
          switch (val?.provider?.value) {
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
                    aws: {
                      awsAccountId: val.awsAccountId,
                    },
                    cloudProviderName: validateCloudProvider(
                      val.provider.value
                    ),
                  },
                });
              }

              return api.createProviderSecret({
                secret: {
                  displayName: val.displayName,
                  metadata: {
                    name: val.name,
                  },
                  aws: {
                    accessKey: val.accessKey,
                    secretKey: val.accessSecret,
                  },
                  cloudProviderName: validateCloudProvider(val.provider.value),
                },
              });

            default:
              throw new Error('invalid provider');
          }
        };

        const updateProvider = async () => {
          if (!isUpdate) {
            throw new Error("data can't be null");
          }

          switch (val?.provider?.value) {
            case 'aws':
              return api.updateProviderSecret({
                secret: {
                  cloudProviderName: props.data.cloudProviderName,
                  displayName: val.displayName,
                  metadata: {
                    name: parseName(props.data, true),
                  },
                  aws: { ...props.data.aws },
                },
              });
            default:
              throw new Error('invalid provider');
          }
        };

        try {
          if (!isUpdate) {
            const { errors: e } = await addProvider();
            if (e) {
              throw e[0];
            }
            toast.success('provider secret created successfully');
          } else {
            const { errors: e } = await updateProvider();
            if (e) {
              throw e[0];
            }
          }
          reloadPage();
          setVisible(false);
          resetValues();
        } catch (err) {
          handleError(err);
        }
      },
    });

  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content>
        <div className="flex flex-col gap-2xl">
          {isUpdate && (
            <Chips.Chip
              {...{
                item: { id: parseName(props.data) },
                label: parseName(props.data),
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
          {!isUpdate && (
            <IdSelector
              name={values.displayName}
              resType="providersecret"
              onChange={(id) => {
                handleChange('name')({ target: { value: id } });
              }}
            />
          )}
          {!isUpdate && (
            <Select
              valueRender={valueRender}
              error={!!errors.provider}
              message={errors.provider}
              value={values.provider}
              label="Provider"
              onChange={(value) => {
                handleChange('provider')(dummyEvent(value));
              }}
              options={async () => providers}
            />
          )}

          {!isUpdate && values?.provider?.value === 'aws' && (
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
          content={isUpdate ? 'Update' : 'Add'}
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleProvider = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>
        {isUpdate ? 'Edit cloud provider' : 'Add new cloud provider'}
      </Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleProvider;
