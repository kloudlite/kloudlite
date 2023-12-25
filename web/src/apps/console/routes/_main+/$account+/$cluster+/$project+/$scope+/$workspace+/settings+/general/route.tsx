import { CopySimple } from '@jengaicons/react';
import { useOutletContext } from '@remix-run/react';
import { useEffect } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/console/components/common-console-components';
import SubNavAction from '~/console/components/sub-nav-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IWorkspace } from '~/console/server/gql/queries/workspace-queries';
import { ConsoleApiType } from '~/console/server/gql/saved-queries';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import useForm from '~/root/lib/client/hooks/use-form';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IWorkspaceContext } from '../../_layout';

export const updateWorkspace = async ({
  api,
  data,
}: {
  api: ConsoleApiType;
  data: ExtractNodeType<IWorkspace>;
}) => {
  try {
    const { errors: e } = await api.updateWorkspace({
      env: {
        displayName: data.displayName,
        metadata: data.metadata,
      },
    });
    if (e) {
      throw e[0];
    }
  } catch (err) {
    handleError(err);
  }
};

const WorkspaceSettingGeneral = () => {
  const { workspace, project, account, cluster } =
    useOutletContext<IWorkspaceContext>();

  const { setHasChanges, resetAndReload } = useUnsavedChanges();

  const api = useConsoleApi();

  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  const { values, handleChange, submit, isLoading, resetValues } = useForm({
    initialValues: {
      displayName: workspace.displayName,
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required('Name is required.'),
    }),
    onSubmit: async (val) => {
      await updateWorkspace({
        api,
        data: { ...workspace, displayName: val.displayName },
      });
      resetAndReload();
    },
  });

  useEffect(() => {
    setHasChanges(values.displayName !== workspace.displayName);
  }, [values]);

  useEffect(() => {
    resetValues();
  }, [workspace]);

  return (
    <>
      <SubNavAction deps={[values]}>
        {values.displayName !== workspace.displayName && (
          <>
            <Button
              content="Discard"
              variant="basic"
              onClick={() => {
                resetValues();
              }}
            />
            <Button
              content="Save changes"
              variant="primary"
              loading={isLoading}
              onClick={() => {
                if (!isLoading) submit();
              }}
            />
          </>
        )}
      </SubNavAction>
      <Box title="General">
        <TextInput
          label="Workspace name"
          value={values.displayName}
          onChange={handleChange('displayName')}
        />
        <div className="flex flex-row items-center gap-3xl">
          <div className="flex-1">
            <TextInput
              label="Workspace URL"
              value={`${consoleBaseUrl}/${parseName(account)}/${parseName(
                cluster
              )}/${parseName(project)}/workspace/${parseName(workspace)}`}
              message="This is your URL namespace within Kloudlite"
              disabled
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    aria-label="Copy"
                    onClick={() =>
                      copy(
                        `${consoleBaseUrl}/${parseName(account)}/${parseName(
                          cluster
                        )}/${parseName(workspace)}`
                      )
                    }
                    className="outline-none hover:bg-surface-basic-hovered active:bg-surface-basic-active rounded text-text-default"
                    tabIndex={-1}
                  >
                    <CopySimple size={16} />
                  </button>
                </div>
              }
            />
          </div>
          <div className="flex-1">
            <TextInput
              value={parseName(workspace)}
              label="Workspace ID"
              message="Used when interacting with the Kloudlite API"
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    aria-label="Copy"
                    onClick={() => copy(parseName(workspace))}
                    className="outline-none hover:bg-surface-basic-hovered active:bg-surface-basic-active rounded text-text-default"
                    tabIndex={-1}
                  >
                    <CopySimple size={16} />
                  </button>
                </div>
              }
              disabled
            />
          </div>
        </div>
      </Box>

      <DeleteContainer title="Delete Workspace" action={() => {}}>
        Permanently remove your Workspace and all of its contents from the
        Kloudlite platform. This action is not reversible — please continue with
        caution.
      </DeleteContainer>
    </>
  );
};
export default WorkspaceSettingGeneral;
