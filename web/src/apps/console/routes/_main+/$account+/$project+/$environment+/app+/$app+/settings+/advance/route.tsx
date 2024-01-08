import { useNavigate, useOutletContext } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { DeleteContainer } from '~/console/components/common-console-components';
import { useAppState } from '~/console/page-components/app-states';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { parseName } from '~/console/server/r-utils/common';
import DeleteDialog from '~/console/components/delete-dialog';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { toast } from '~/components/molecule/toast';
import { handleError } from '~/root/lib/utils/common';
import { IAppContext } from '../../route';

const SettingAdvance = () => {
  const { app, setApp } = useAppState();
  const { environment, project } = useOutletContext<IAppContext>();
  const [deleteApp, setDeleteApp] = useState(false);
  const reload = useReload();
  const api = useConsoleApi();
  const navigate = useNavigate();

  const { values, submit } = useForm({
    initialValues: {
      name: parseName(app),
      displayName: app.displayName,
      description: app.metadata?.annotations?.[keyconstants.description] || '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
      description: Yup.string(),
    }),

    onSubmit: async (val) => {
      setApp((a) => {
        return {
          ...a,
          metadata: {
            ...a.metadata,
            name: val.name,
            namespace: environment.spec?.targetNamespace,
            annotations: {
              ...(a.metadata?.annotations || {}),
              [keyconstants.description]: val.description,
            },
          },
          displayName: val.displayName,
        };
      });
    },
  });

  useEffect(() => {
    submit();
  }, [values]);

  return (
    <>
      <div className="rounded border border-border-default bg-surface-basic-default shadow-button flex flex-col">
        <div className="flex flex-col gap-3xl p-3xl">
          <div className="text-text-strong headingLg">Transfer</div>
          <div className="bodyMd text-text-default">
            Move your app to a different environment seamlessly, avoiding any
            downtime or disruptions to workflows.
          </div>
        </div>
        <div className="bg-surface-basic-subdued p-3xl flex flex-row justify-end">
          <Button variant="basic" content="Transfer" />
        </div>
      </div>
      <DeleteContainer
        title="Delete Application"
        action={async () => {
          setDeleteApp(true);
        }}
      >
        Permanently remove your application and all of its contents from the
        “Lobster Early” project. This action is not reversible, so please
        continue with caution.
      </DeleteContainer>
      <DeleteDialog
        resourceName={parseName(app)}
        resourceType="app"
        show={deleteApp}
        setShow={setDeleteApp}
        onSubmit={async () => {
          try {
            const { errors } = await api.deleteApp({
              appName: parseName(app),
              envName: parseName(environment),
              projectName: parseName(project),
            });

            if (errors) {
              throw errors[0];
            }
            reload();
            toast.success(`App deleted successfully`);
            setDeleteApp(false);
            navigate(`../`);
          } catch (err) {
            handleError(err);
          }
        }}
      />
    </>
  );
};
export default SettingAdvance;
