import { CopySimple } from '@jengaicons/react';
import { useNavigate, useOutletContext } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/console/components/common-console-components';
import SubNavAction from '~/console/components/sub-nav-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ConsoleApiType } from '~/console/server/gql/saved-queries';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import useForm from '~/root/lib/client/hooks/use-form';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IEnvironment } from '~/console/server/gql/queries/environment-queries';
import DeleteDialog from '~/console/components/delete-dialog';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { IEnvironmentContext } from '../../_layout';

export const updateEnvironment = async ({
  api,
  data,
  project,
}: {
  project: string;
  api: ConsoleApiType;
  data: ExtractNodeType<IEnvironment>;
}) => {
  try {
    const { errors: e } = await api.updateEnvironment({
      projectName: project,
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

const EnvironmentSettingsGeneral = () => {
  const { environment, project, account } =
    useOutletContext<IEnvironmentContext>();

  const { setHasChanges, resetAndReload } = useUnsavedChanges();
  const [deleteEnvironment, setDeleteEnvironment] = useState(false);

  const api = useConsoleApi();
  const reload = useReload();
  const navigate = useNavigate();

  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  const { values, handleChange, submit, isLoading, resetValues } = useForm({
    initialValues: {
      displayName: environment.displayName,
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required('Name is required.'),
    }),
    onSubmit: async (val) => {
      await updateEnvironment({
        project: parseName(project),
        api,
        data: { ...environment, displayName: val.displayName },
      });
      resetAndReload();
    },
  });

  useEffect(() => {
    setHasChanges(values.displayName !== environment.displayName);
  }, [values]);

  useEffect(() => {
    resetValues();
  }, [environment]);

  return (
    <>
      <SubNavAction deps={[values]}>
        {values.displayName !== environment.displayName && (
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
          label="Environment name"
          value={values.displayName}
          onChange={handleChange('displayName')}
        />
        <div className="flex flex-row items-center gap-3xl">
          <div className="flex-1">
            <TextInput
              label="Environment URL"
              value={`${consoleBaseUrl}/${parseName(account)}/${parseName(
                project
              )}/${parseName(environment)}`}
              message="This is your URL namespace within Kloudlite"
              disabled
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    aria-label="Copy"
                    onClick={() =>
                      copy(
                        `${consoleBaseUrl}/${parseName(account)}/${parseName(
                          environment
                        )}`
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
              value={parseName(environment)}
              label="Environment ID"
              message="Used when interacting with the Kloudlite API"
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    aria-label="Copy"
                    onClick={() => copy(parseName(environment))}
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

      <DeleteContainer
        title="Delete Application"
        action={async () => {
          setDeleteEnvironment(true);
        }}
      >
        Permanently remove your environment and all of its contents from the “
        {environment.displayName}” environment. This action is not reversible,
        so please continue with caution.
      </DeleteContainer>
      <DeleteDialog
        resourceName={parseName(environment)}
        resourceType="environment"
        show={deleteEnvironment}
        setShow={setDeleteEnvironment}
        onSubmit={async () => {
          try {
            const { errors } = await api.deleteEnvironment({
              envName: parseName(environment),
              projectName: parseName(project),
            });

            if (errors) {
              throw errors[0];
            }
            reload();
            toast.success(`Environment deleted successfully`);
            setDeleteEnvironment(false);
            navigate(
              `/${parseName(account)}/${parseName(project)}/environments/`
            );
          } catch (err) {
            handleError(err);
          }
        }}
      />
    </>
  );
};
export default EnvironmentSettingsGeneral;
