import { useNavigate, useOutletContext, useParams } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { DeleteContainer } from '~/console/components/common-console-components';
import { useAppState } from '~/console/page-components/app-states';
import { parseName } from '~/console/server/r-utils/common';
import DeleteDialog from '~/console/components/delete-dialog';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { toast } from '~/components/molecule/toast';
import { handleError } from '~/root/lib/utils/common';
import Wrapper from '~/console/components/wrapper';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { IAppContext } from '../../route';

const SettingAdvance = () => {
  const { app } = useAppState();
  const { environment, project } = useOutletContext<IAppContext>();
  const [deleteApp, setDeleteApp] = useState(false);
  const reload = useReload();
  const api = useConsoleApi();
  const navigate = useNavigate();
  const { setPerformAction, hasChanges, loading } = useUnsavedChanges();
  const { account } = useParams();

  return (
    <div>
      <Wrapper
        secondaryHeader={{
          title: 'Advance',
          action: hasChanges && (
            <div className="flex flex-row items-center gap-lg">
              <Button
                disabled={loading}
                variant="basic"
                content="Discard changes"
                onClick={() => setPerformAction('discard-changes')}
              />
              <Button
                disabled={loading}
                content={loading ? 'Committing changes.' : 'View changes'}
                loading={loading}
                onClick={() => setPerformAction('view-changes')}
              />
            </div>
          ),
        }}
      >
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
              navigate(
                `/${account}/${parseName(project)}/${parseName(
                  environment
                )}/apps/`
              );
            } catch (err) {
              handleError(err);
            }
          }}
        />
      </Wrapper>
    </div>
  );
};
export default SettingAdvance;
