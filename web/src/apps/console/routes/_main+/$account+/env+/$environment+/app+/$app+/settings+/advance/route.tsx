import { useNavigate, useOutletContext } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '@kloudlite/design-system/atoms/button';
import { toast } from '@kloudlite/design-system/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/console/components/common-console-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Wrapper from '~/console/components/wrapper';
import YamlEditorOverlay from '~/console/components/yaml-editor-overlay';
import { useAppState } from '~/console/page-components/app-states';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName } from '~/console/server/r-utils/common';
import { getAppIn } from '~/console/server/r-utils/resource-getter';
import { useReload } from '~/lib/client/helpers/reloader';
import { useUnsavedChanges } from '~/lib/client/hooks/use-unsaved-changes';
import { handleError } from '~/lib/utils/common';
import { IAppContext } from '../../_layout';

const SettingAdvance = () => {
  const { app, readOnlyApp } = useAppState();
  const { environment } = useOutletContext<IAppContext>();
  const [deleteApp, setDeleteApp] = useState(false);
  const reload = useReload();
  const api = useConsoleApi();
  const navigate = useNavigate();
  const { setPerformAction, hasChanges, loading } = useUnsavedChanges();

  const [showDialog, setShowDialog] = useState(false);

  const updateApp = async ({ spec }: { spec: any }) => {
    try {
      const { errors: e } = await api.updateApp({
        app: {
          ...getAppIn({
            ...readOnlyApp,
            spec,
          }),
        },
        envName: parseName(environment),
      });
      if (e) {
        throw e[0];
      }
      toast.success('App updated successfully');
      // reload();
      return true;
    } catch (e) {
      toast.error('error while updating app');
      return false;
    }
  };

  return (
    <div>
      <Wrapper
        secondaryHeader={{
          title: 'Advanced',
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
        <Box title="Edit App as yaml">
          <div className="flex flex-col gap-3xl">
            <span className="bodyMd text-text-default">
              You are allowed to edit your “{app.displayName}” as yaml. This
              action will open an yaml editor where you can edit in yaml and
              commit your changes.
            </span>
            <div>
              <Button content="Edit Yaml" onClick={() => setShowDialog(true)} />
              <YamlEditorOverlay
                item={readOnlyApp}
                showDialog={showDialog}
                setShowDialog={() => {
                  setShowDialog(false);
                }}
                onCommit={updateApp}
              />
            </div>
          </div>
        </Box>

        {/* <div className="rounded border border-border-default bg-surface-basic-default shadow-button flex flex-col"> */}
        {/*   <div className="flex flex-col gap-3xl p-3xl"> */}
        {/*     <div className="text-text-strong headingLg">Transfer</div> */}
        {/*     <div className="bodyMd text-text-default"> */}
        {/*       Move your app to a different environment seamlessly, avoiding any */}
        {/*       downtime or disruptions to workflows. */}
        {/*     </div> */}
        {/*   </div> */}
        {/*   <div className="bg-surface-basic-subdued p-3xl flex flex-row justify-end"> */}
        {/*     <Button variant="basic" content="Transfer" /> */}
        {/*   </div> */}
        {/* </div> */}
        <DeleteContainer
          title="Delete Application"
          action={async () => {
            setDeleteApp(true);
          }}
        >
          Permanently remove your application and all of its contents from the “
          {app.displayName}” project. This action is not reversible, so please
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
              });

              if (errors) {
                throw errors[0];
              }
              reload();
              toast.success(`App deleted successfully`);
              setDeleteApp(false);
              navigate(`../../../apps/`);
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
