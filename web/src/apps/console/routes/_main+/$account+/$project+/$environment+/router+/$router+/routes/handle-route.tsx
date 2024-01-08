/* eslint-disable react/destructuring-assignment */
import { useParams } from '@remix-run/react';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IDialogBase } from '~/console/components/types.d';
import { IRouter, IRouters } from '~/console/server/gql/queries/router-queries';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import Select from '~/components/atoms/select';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { useMapper } from '~/components/utils';
import { NN } from '~/root/lib/types/common';
import { TextInput } from '~/components/atoms/input';
import { useState } from 'react';
import { IApp } from '~/console/server/gql/queries/app-queries';

type IDialog = IDialogBase<
  NN<ExtractNodeType<IRouters>['spec']['routes']>[number]
> & { router?: IRouter };

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { project: projectName, environment: envName } = useParams();
  const [selectedApp, setSelectedApp] = useState<IApp>();

  const {
    data,
    isLoading: appLoading,
    error: appLoadingError,
  } = useCustomSwr('/apps', async () => {
    if (!projectName || !envName) {
      throw new Error('Project and Environment is required!.');
    }
    return api.listApps({
      projectName,
      envName,
    });
  });

  const { values, errors, handleSubmit, handleChange, isLoading, resetValues } =
    useForm({
      initialValues: {
        path: '',
        app: '',
        port: '',
      },
      validationSchema: Yup.object({
        path: Yup.string().required(),
        app: Yup.string().required(),
        port: Yup.string().required(),
      }),

      onSubmit: async (val) => {
        const { router } = props;
        if (!projectName || !envName || !router || !router.metadata?.name) {
          throw new Error('Project, Router and Environment is required!.');
        }
        try {
          if (!isUpdate) {
            const { errors: e } = await api.updateRouter({
              envName,
              projectName,
              router: {
                displayName: router.displayName,
                spec: {
                  domains: router.spec.domains,
                  routes: [
                    ...(router.spec.routes || []),
                    {
                      path: val.path,
                      app: val.app,
                      port: parseInt(val.port, 10),
                    },
                  ],
                },
                metadata: {
                  ...router.metadata,
                },
              },
            });
            if (e) {
              throw e[0];
            }
            toast.success('Route created successfully');
          } else {
            //
          }
          reloadPage();
          setVisible(false);
          resetValues();
        } catch (err) {
          handleError(err);
        }
      },
    });

  const apps = useMapper(parseNodes(data), (val) => ({
    label: val.displayName,
    value: parseName(val),
    app: val,
    render: () => val.displayName,
  }));

  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content className="flex flex-col gap-3xl">
        <TextInput
          label="Path"
          size="lg"
          value={values.path}
          onChange={handleChange('path')}
          error={!!errors.path}
          message={errors.path}
        />
        <Select
          size="lg"
          label="App"
          value={{ label: '', value: values.app }}
          options={async () => [...apps]}
          onChange={(val) => {
            handleChange('app')(dummyEvent(val.value));
            setSelectedApp(val.app);
          }}
          error={!!errors.app || !!appLoadingError}
          message={appLoadingError ? 'Error fetching apps.' : errors.app}
          loading={appLoading}
        />
        <Select
          size="lg"
          label="Port"
          disabled={!values.app}
          value={{ label: '', value: values.port }}
          options={async () => [
            ...(selectedApp?.spec.services?.map((svc) => ({
              label: `${svc.port}`,
              value: `${svc.port}`,
            })) || []),
          ]}
          onChange={(val) => {
            handleChange('port')(dummyEvent(val.value));
          }}
          error={!!errors.port}
          message={errors.port}
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

const HandleRoute = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;
  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>{isUpdate ? 'Edit route' : 'Add Route'}</Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleRoute;
