import { useOutletContext } from '@remix-run/react';
import { PasswordInput, TextInput } from '~/components/atoms/input';
import * as Popup from '~/components/molecule/popup';
import * as SelectInput from '~/components/atoms/select';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { IdSelector, idTypes } from '~/console/components/id-selector';
import { useAPIClient } from '~/console/server/utils/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { getSecretRef } from '~/console/server/r-urils/secret-ref';
import {
  getMetadata,
  parseDisplaynameFromAnn,
  parseFromAnn,
  parseName,
} from '~/console/server/r-urils/common';
import { keyconstants } from '~/console/server/r-urils/key-constants';
import * as Chips from '~/components/atoms/chips';
import { toast } from '~/components/molecule/toast';
import { useEffect, useState } from 'react';

const HandleProvider = ({ show, setShow, onSubmit }) => {
  const api = useAPIClient();
  const reloadPage = useReload();
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
        if (show.type === 'add') {
          const { errors: e } = await api.createProviderSecret({
            secret: getSecretRef({
              metadata: getMetadata({
                name: val.name,
                annotations: {
                  [keyconstants.displayName]: val.displayName,
                  [keyconstants.provider]: val.provider,
                  [keyconstants.author]: user.name,
                },
              }),
              stringData: {
                accessKey: val.accessKey,
                accessSecret: val.accessSecret,
              },
            }),
          });
          if (e) {
            throw e[0];
          }
          toast.success('provider secret created successfully');
        } else {
          console.log(
            'provider',
            parseFromAnn(show.data, keyconstants.provider)
          );
          const { errors: e } = await api.updateProviderSecret({
            secret: getSecretRef({
              metadata: getMetadata({
                name: parseName(show.data),
                annotations: {
                  [keyconstants.displayName]: val.displayName,
                  [keyconstants.provider]: parseFromAnn(
                    show.data,
                    keyconstants.provider
                  ),
                  [keyconstants.author]: user.name,
                },
              }),
              stringData: {
                accessKey: val.accessKey,
                accessSecret: val.accessSecret,
              },
            }),
          });
          if (e) {
            throw e[0];
          }
        }
        reloadPage();
        setShow(false);
        resetValues();
      } catch (err) {
        toast.error(err.message);
      }
    },
  });

  useEffect(() => {
    if (show?.type === 'edit') {
      setValues({
        displayName: parseDisplaynameFromAnn(show.data),
        accessSecret: show.data?.stringData?.accessSecret || '',
        accessKey: show.data?.stringData?.accessKey || '',
      });
      setValidationSchema(
        Yup.object({
          displayName: Yup.string().trim().required(),
          accessSecret: Yup.string().trim().required(),
          accessKey: Yup.string().trim().required(),
        })
      );
    }
  }, [show]);

  return (
    <Popup.PopupRoot
      show={show}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }
        setShow(e);
      }}
    >
      <Popup.Header>
        {show.type === 'add' ? 'Add new cloud provider' : 'Edit cloud provider'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            {show.type === 'edit' && (
              <Chips.Chip
                {...{
                  item: { id: parseName(show.data) },
                  label: parseName(show.data),
                  prefix: 'Id:',
                  disabled: true,
                  type: Chips.ChipType.BASIC,
                }}
              />
            )}
            <TextInput
              label="Name"
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={!!errors.displayName}
              value={values.displayName}
              name="provider-secret-name"
            />
            {show.type === 'add' && (
              <IdSelector
                name={values.displayName}
                resType={idTypes.providersecret}
                onChange={(id) => {
                  handleChange('name')({ target: { value: id } });
                }}
              />
            )}
            {show.type === 'add' && (
              <SelectInput.Select
                error={!!errors.provider}
                message={errors.provider}
                value={values.provider}
                label="Provider"
                onChange={(provider) => {
                  handleChange('provider')({ target: { value: provider } });
                }}
              >
                <SelectInput.Option value="aws">
                  Amazon Web Services
                </SelectInput.Option>
              </SelectInput.Select>
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
            content={show.type === 'add' ? 'Add' : 'Update'}
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.PopupRoot>
  );
};

export default HandleProvider;
