/* eslint-disable react/destructuring-assignment */
// import { useParams } from '@remix-run/react';
import { useEffect, useRef } from 'react';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import {
  ExtractNodeType,
  parseName,
  validateImagePullSecretFormat,
} from '~/console/server/r-utils/common';
import { useReload } from '~/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/lib/client/hooks/use-form';
import Yup from '~/lib/server/helpers/yup';
import { handleError } from '~/lib/utils/common';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { IDialogBase } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { NameIdView } from '~/console/components/name-id-view';
import { IImagePullSecrets } from '~/console/server/gql/queries/image-pull-secrets-queries';
import { PasswordInput, TextArea, TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';

type IDialog = IDialogBase<ExtractNodeType<IImagePullSecrets>>;

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const formats = [
    {
      label: 'Params',
      value: 'params',
      render: () => (
        <div className="flex flex-row gap-lg items-center">
          <div>Params</div>
        </div>
      ),
    },
    {
      label: 'Docker Config Json',
      value: 'dockerConfigJson',
      render: () => (
        <div className="flex flex-row gap-lg items-center">
          <div>Docker Config Json</div>
        </div>
      ),
    },
  ];

  // const { environment: envName } = useParams();

  const { values, errors, handleSubmit, handleChange, isLoading, resetValues } =
    useForm({
      initialValues: isUpdate
        ? {
            name: parseName(props.data),
            displayName: props.data.displayName,
            registryUsername: props.data.registryUsername,
            registryPassword: props.data.registryPassword,
            registryURL: props.data.registryURL,
            dockerConfigJson: props.data.dockerConfigJson,
            format: props.data.format,
            domains: [],
            isNameError: false,
          }
        : {
            name: '',
            displayName: '',
            registryUsername: '',
            registryPassword: '',
            registryURL: '',
            dockerConfigJson: '',
            format: formats[0].value,
            domains: [],
            isNameError: false,
          },
      validationSchema: Yup.object({
        displayName: Yup.string().required(),
        name: Yup.string().required(),
        registryUsername: Yup.string().when(['format'], ([format], schema) => {
          if (format === 'dockerConfigJson') {
            return schema.notRequired();
          }
          return schema.required();
        }),
        registryPassword: Yup.string().when(['format'], ([format], schema) => {
          if (format === 'dockerConfigJson') {
            return schema.notRequired();
          }
          return schema.required();
        }),
        registryURL: Yup.string().when(['format'], ([format], schema) => {
          if (format === 'dockerConfigJson') {
            return schema.notRequired();
          }
          return schema.required();
        }),
        dockerConfigJson: Yup.string().when(['format'], ([format], schema) => {
          if (format === 'params') {
            return schema.notRequired();
          }
          return schema.required();
        }),
      }),

      onSubmit: async (val) => {
        const addImagePullSecret = async () => {
          console.log('format', val.format);

          switch (val?.format) {
            case 'params':
              return api.createImagePullSecret({
                pullSecret: {
                  displayName: val.displayName,
                  metadata: {
                    name: val.name,
                  },
                  registryUsername: val.registryUsername,
                  registryPassword: val.registryPassword,
                  registryURL: val.registryURL,
                  format: validateImagePullSecretFormat(val.format),
                },
              });
            case 'dockerConfigJson':
              return api.createImagePullSecret({
                pullSecret: {
                  displayName: val.displayName,
                  metadata: {
                    name: val.name,
                  },
                  dockerConfigJson: val.dockerConfigJson,
                  format: validateImagePullSecretFormat(val.format),
                },
              });

            default:
              throw new Error('invalid provider');
          }
        };

        const updateImagePullSecret = async () => {
          if (!isUpdate) {
            throw new Error("data can't be null");
          }

          switch (val?.format) {
            case 'params':
              return api.updateImagePullSecret({
                pullSecret: {
                  displayName: val.displayName,
                  metadata: {
                    name: val.name,
                  },
                  registryUsername: val.registryUsername,
                  registryPassword: val.registryPassword,
                  registryURL: val.registryURL,
                  format: validateImagePullSecretFormat(val.format),
                },
              });
            case 'dockerConfigJson':
              return api.updateImagePullSecret({
                pullSecret: {
                  displayName: val.displayName,
                  metadata: {
                    name: val.name,
                  },
                  dockerConfigJson: val.dockerConfigJson,
                  format: validateImagePullSecretFormat(val.format),
                },
              });

            default:
              throw new Error('invalid provider');
          }
        };

        try {
          if (!isUpdate) {
            const { errors: e } = await addImagePullSecret();
            if (e) {
              throw e[0];
            }
            toast.success('Image pull secrets created successfully');
          } else {
            const { errors: e } = await updateImagePullSecret();
            if (e) {
              throw e[0];
            }
            if (e) {
              throw e[0];
            }
            toast.success('Image pull secrets updated successfully');
          }
          reloadPage();
          setVisible(false);
          resetValues();
        } catch (err) {
          handleError(err);
        }
      },
    });

  const nameIDRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    nameIDRef.current?.focus();
  }, [nameIDRef]);

  useEffect(() => {
    console.log('fff', values.format, values.dockerConfigJson);
    console.log(errors);
  }, [values.format, values.dockerConfigJson, errors]);

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
      <Popup.Content className="flex flex-col justify-start gap-3xl">
        <Select
          error={!!errors.format}
          message={errors.format}
          value={values.format}
          label="Format"
          onChange={(_, value) => {
            handleChange('format')(dummyEvent(value));
          }}
          options={async () => formats}
          disabled={isUpdate}
        />

        <NameIdView
          ref={nameIDRef}
          // resType="router"
          resType="environment"
          label="Name"
          placeholder="Enter image pull secret name"
          displayName={values.displayName}
          name={values.name}
          errors={errors.name}
          handleChange={handleChange}
          nameErrorLabel="isNameError"
          isUpdate={isUpdate}
        />

        {values.format === 'params' && (
          <>
            <TextInput
              size="lg"
              label="Registry url"
              placeholder="Enter registry url"
              value={values.registryURL}
              onChange={handleChange('registryURL')}
              error={!!errors.registryURL}
              message={errors.registryURL}
            />
            <TextInput
              size="lg"
              label="Registry username"
              placeholder="Enter registry username"
              value={values.registryUsername}
              onChange={handleChange('registryUsername')}
              error={!!errors.registryUsername}
              message={errors.registryUsername}
            />
            <PasswordInput
              size="lg"
              label="Registry password"
              placeholder="Enter registry password"
              value={values.registryPassword}
              onChange={handleChange('registryPassword')}
              error={!!errors.registryPassword}
              message={errors.registryPassword}
            />
          </>
        )}

        {values?.format === 'dockerConfigJson' && (
          <TextArea
            placeholder="Enter docker config json"
            label="Docker Config JSON"
            value={values.dockerConfigJson}
            onChange={handleChange('dockerConfigJson')}
            resize={false}
            rows="6"
            error={!!errors.dockerConfigJson}
            message={errors.dockerConfigJson}
          />
        )}
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button content="Cancel" variant="basic" closable />
        <Popup.Button
          loading={isLoading}
          type="submit"
          content={!isUpdate ? 'Add' : 'Update'}
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleImagePullSecret = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      createTitle="Create image pull secret"
      updateTitle="Update image pull secret"
      root={Root}
    />
  );
};
export default HandleImagePullSecret;
