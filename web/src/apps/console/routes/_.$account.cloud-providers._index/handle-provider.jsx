import { useOutletContext } from '@remix-run/react';
import { PasswordInput, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import Select from '~/components/atoms/select';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { IdSelector } from '~/console/components/id-selector';
import { useReload } from '~/root/lib/client/helpers/reloader';
import {
  parseDisplaynameFromAnn,
  parseName,
} from '~/console/server/r-urils/common';
import { keyconstants } from '~/console/server/r-urils/key-constants';
import * as Chips from '~/components/atoms/chips';
import { toast } from '~/components/molecule/toast';
import { useEffect, useState } from 'react';
import { handleError } from '~/root/lib/utils/common';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { validateCloudProvider } from '~/root/src/generated/r-types/utils';

const HandleProvider = ({ show, setShow }) => {
  const api = useConsoleApi();
  const reloadPage = useReload();
  // @ts-ignore
  const { user } = useOutletContext();

  const [validationSchema, setValidationSchema] = useState(
    Yup.object({
      displayName: Yup.string().required(),
      name: Yup.string().required(),
      provider: Yup.string().required(),
      accessKey: Yup.string().required(),
      accessSecret: Yup.string().required(),
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
    },
    validationSchema,

    onSubmit: async (val) => {
      try {
        if (show?.type === 'add') {
          console.log(val);
          const { errors: e } = await api.createProviderSecret({
            secret: {
              metadata: show.data.metadata,
              stringData: {
                accessKey: val.accessKey,
                accessSecret: val.accessSecret,
              },
              cloudProviderName: validateCloudProvider(val.provider),
            },
          });
          if (e) {
            throw e[0];
          }
          toast.success('provider secret created successfully');
        } else {
          const { errors: e } = await api.updateProviderSecret({
            secret: {
              metadata: {
                name: parseName(show.data),
                annotations: {
                  [keyconstants.displayName]: val.displayName,
                  [keyconstants.author]: user.name,
                },
              },
              stringData: {
                accessKey: val.accessKey,
                accessSecret: val.accessSecret,
              },
              cloudProviderName: val.provider,
            },
          });
          if (e) {
            throw e[0];
          }
        }
        reloadPage();
        setShow(false);
        resetValues();
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    if (show?.type === 'edit') {
      setValues((v) => ({
        ...v,
        displayName: parseDisplaynameFromAnn(show.data),
        accessSecret: show.data?.stringData?.accessSecret || '',
        accessKey: show.data?.stringData?.accessKey || '',
      }));
      setValidationSchema(
        // @ts-ignore
        Yup.object({
          displayName: Yup.string().trim().required(),
          accessSecret: Yup.string().trim().required(),
          accessKey: Yup.string().trim().required(),
          provider: Yup.string().required(),
        })
      );
    }
  }, [show]);

  return (
    <Popup.Root
      show={show}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }
        setShow(e);
      }}
    >
      <Popup.Header>
        {show?.type === 'add'
          ? 'Add new cloud provider'
          : 'Edit cloud provider'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            {show?.type === 'edit' && (
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
            {show?.type === 'add' && (
              <IdSelector
                name={values.displayName}
                resType="providersecret"
                onChange={(id) => {
                  handleChange('name')({ target: { value: id } });
                }}
              />
            )}
            {show?.type === 'add' && (
              <Select.Root
                error={!!errors.provider}
                message={errors.provider}
                value={values.provider}
                label="Provider"
                onChange={(provider) => {
                  handleChange('provider')({ target: { value: provider } });
                }}
              >
                <Select.Option value="aws">Amazon Web Services</Select.Option>
              </Select.Root>
            )}
            <PasswordInput
              name="accessKey"
              onChange={handleChange('accessKey')}
              error={!!errors.accessKey}
              message={errors.accessKey}
              value={values.accessKey}
              label="Access Key ID"
            />
            <PasswordInput
              name="accessSecret"
              label="Access Key Secret"
              onChange={handleChange('accessSecret')}
              error={!!errors.accessSecret}
              message={errors.accessSecret}
              value={values.accessSecret}
            />
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button content="Cancel" variant="basic" closable />
          <Popup.Button
            loading={isLoading}
            type="submit"
            content={show?.type === 'add' ? 'Add' : 'Update'}
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

export default HandleProvider;
