/* eslint-disable react/destructuring-assignment */
import { ReactNode } from 'react';
import { PasswordInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { NameIdView } from '~/console/components/name-id-view';
import { IDialogBase } from '~/console/components/types.d';
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
          <div>Amazon Web Services</div>
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
            provider: props.data.cloudProviderName as string,
            accessKey: '',
            secretKey: '',
            isNameError: false,
          }
        : {
            displayName: '',
            name: '',
            provider: providers[0].value,
            accessKey: '',
            secretKey: '',
            isNameError: false,
          },
      validationSchema: isUpdate
        ? Yup.object({
            displayName: Yup.string().required(),
            name: Yup.string().required(),
          })
        : Yup.object({
            displayName: Yup.string().required(),
            name: Yup.string().required(),
            provider: Yup.string().required(),
            accessKey: Yup.string().test(
              'provider',
              'access key is required',
              function (item) {
                return (
                  // @ts-ignores
                  // eslint-disable-next-line react/no-this-in-sfc
                  this.parent.provider &&
                  // eslint-disable-next-line react/no-this-in-sfc
                  this.parent.provider === 'aws' &&
                  item
                );
              }
            ),
            secretKey: Yup.string().test(
              'provider',
              'secret key is required',
              function (item) {
                return (
                  // @ts-ignores
                  // eslint-disable-next-line react/no-this-in-sfc
                  this.parent.provider &&
                  // eslint-disable-next-line react/no-this-in-sfc
                  this.parent.provider === 'aws' &&
                  item
                );
              }
            ),
          }),

      onSubmit: async (val) => {
        const addProvider = async () => {
          switch (val?.provider) {
            case 'aws':
              return api.createProviderSecret({
                secret: {
                  displayName: val.displayName,
                  metadata: {
                    name: val.name,
                  },
                  aws: {
                    secretKey: val.secretKey,
                    accessKey: val.accessKey,
                  },
                  cloudProviderName: validateCloudProvider(val.provider),
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

          switch (val?.provider) {
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
    <Popup.Form
      onSubmit={(e) => {
        if (!values.isNameError) {
          handleSubmit(e);
        } else {
          e.preventDefault();
        }
      }}
    >
      <Popup.Content>
        <div className="flex flex-col gap-2xl">
          <NameIdView
            resType="providersecret"
            displayName={values.displayName}
            name={values.name}
            label="Name"
            placeholder="Enter cloud provider name"
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
            isUpdate={isUpdate}
          />
          {!isUpdate && (
            <Select
              valueRender={valueRender}
              error={!!errors.provider}
              message={errors.provider}
              value={values.provider}
              label="Provider"
              onChange={(_, value) => {
                handleChange('provider')(dummyEvent(value));
              }}
              options={async () => providers}
            />
          )}

          {!isUpdate && values?.provider === 'aws' && (
            <>
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
            </>
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
