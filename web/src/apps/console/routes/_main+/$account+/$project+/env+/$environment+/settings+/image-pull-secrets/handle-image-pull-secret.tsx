/* eslint-disable react/destructuring-assignment */
import { useParams } from '@remix-run/react';
import { useEffect, useRef, useState } from 'react';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import { useReload } from '~/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/lib/client/hooks/use-form';
import Yup from '~/lib/server/helpers/yup';
import { handleError } from '~/lib/utils/common';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { IDialogBase } from '~/console/components/types.d';
import { IRouters } from '~/console/server/gql/queries/router-queries';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { NameIdView } from '~/console/components/name-id-view';
import Select from '~/components/atoms/select';
import useCustomSwr from '~/lib/client/hooks/use-custom-swr';
import { IDomains } from '~/console/server/gql/queries/domain-queries';
import { useMapper } from '~/components/utils';
import { IImagePullSecrets } from '~/console/server/gql/queries/image-pull-secrets-queries';
import { PasswordInput, TextInput } from '~/components/atoms/input';

type IDialog = IDialogBase<ExtractNodeType<IImagePullSecrets>>;

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { project: projectName, environment: envName } = useParams();

  const { values, errors, handleSubmit, handleChange, isLoading, resetValues } =
    useForm({
      initialValues: isUpdate
        ? {
            name: parseName(props.data),
            displayName: props.data.displayName,
            registryUsername: '',
            registryPassword: '',
            registryURL: '',
            domains: [],
            isNameError: false,
          }
        : {
            name: '',
            displayName: '',
            domains: [],
            isNameError: false,
            registryUsername: '',
            registryPassword: '',
            registryURL: '',
          },
      validationSchema: Yup.object({
        displayName: Yup.string().required(),
        name: Yup.string().required(),
        registryUsername: Yup.string().required(),
        registryPassword: Yup.string().required(),
        registryURL: Yup.string().required(),
        // .test('is-valid', 'invalid domain names', (val) => {
        //   console.log('vals', val);

        //   return val?.every((v) => v.endsWith('.com'));
        // }),
      }),

      onSubmit: async (val) => {
        if (!projectName || !envName) {
          throw new Error('Project, Environment and Domain is required!.');
        }
        try {
          if (!isUpdate) {
            const { errors: e } = await api.createImagePullSecret({
              envName,
              projectName,
              imagePullSecretIn: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },
                registryUsername: val.registryUsername,
                registryPassword: val.registryPassword,
                registryURL: val.registryURL,
                format: 'params',
              },
            });
            if (e) {
              throw e[0];
            }
            toast.success('Image pull secrets created successfully');
          } else {
            // const { errors: e } = await api.updateRouter({
            //   envName,
            //   projectName,
            //   router: {
            //     displayName: val.displayName,
            //     metadata: {
            //       name: val.name,
            //     },
            //     spec: {
            //       ...props.data.spec,
            //       domains: selectedDomains.map((sd) => sd.value),
            //       https: {
            //         enabled: true,
            //       },
            //     },
            //   },
            // });
            // if (e) {
            //   throw e[0];
            // }
            // toast.success('Router updated successfully');
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
        <NameIdView
          ref={nameIDRef}
          resType="router"
          label="Name"
          displayName={values.displayName}
          name={values.name}
          errors={errors.name}
          handleChange={handleChange}
          nameErrorLabel="isNameError"
          isUpdate={isUpdate}
        />
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
