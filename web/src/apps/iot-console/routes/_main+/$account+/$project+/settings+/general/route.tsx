import { CopySimple } from '~/iotconsole/components/icons';
import { useNavigate, useOutletContext } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button } from '@kloudlite/design-system/atoms/button';
import { TextInput } from '@kloudlite/design-system/atoms/input';
import { toast } from '@kloudlite/design-system/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/iotconsole/components/common-console-components';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import { ConsoleApiType } from '~/iotconsole/server/gql/saved-queries';
import { ExtractNodeType, parseName } from '~/iotconsole/server/r-utils/common';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import useForm from '~/root/lib/client/hooks/use-form';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import Wrapper from '~/iotconsole/components/wrapper';
import { useReload } from '~/root/lib/client/helpers/reloader';
import DeleteDialog from '~/iotconsole/components/delete-dialog';
import { IProject } from '~/iotconsole/server/gql/queries/iot-project-queries';
import { IProjectContext } from '../../_layout';

export const updateProject = async ({
  api,
  data,
}: {
  api: ConsoleApiType;
  data: ExtractNodeType<IProject>;
}) => {
  try {
    const { errors: e } = await api.updateIotProject({
      project: {
        displayName: data.displayName,
        name: data.name,
        // metadata: {
        //   name: parseName(data),
        // },
        // spec: {
        //   targetNamespace: data.spec.targetNamespace,
        // },
      },
    });
    if (e) {
      throw e[0];
    }
  } catch (err) {
    handleError(err);
  }
};

const ProjectSettingGeneral = () => {
  const {  account } = useOutletContext<IProjectContext>();

  const { setHasChanges, resetAndReload } = useUnsavedChanges();

  const [deleteProject, setDeleteProject] = useState(false);

  const api = useIotConsoleApi();
  const reload = useReload();
  const navigate = useNavigate();

  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  const { values, handleChange, submit, isLoading, resetValues } = useForm({
    initialValues: {
      displayName: project.displayName,
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required('Name is required.'),
    }),
    onSubmit: async (val) => {
      await updateProject({
        api,
        data: { ... displayName: val.displayName },
      });
      resetAndReload();
    },
  });

  useEffect(() => {
    setHasChanges(values.displayName !== project.displayName);
  }, [values]);

  useEffect(() => {
    resetValues();
  }, [project]);

  return (
    <div>
      <Wrapper
        secondaryHeader={{
          title: 'General',
          action: values.displayName !== project.displayName && (
            <div className="flex flex-row items-center gap-3xl">
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
                onClick={() => {
                  if (!isLoading) submit();
                }}
                loading={isLoading}
              />
            </div>
          ),
        }}
      >
        <Box title="Project details">
          <TextInput
            label="Project name"
            value={values.displayName}
            onChange={handleChange('displayName')}
          />
          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <TextInput
                label="Project URL"
                value={`${consoleBaseUrl}/${parseName(account)}/${
                  project.name
                }`}
                message="This is your URL namespace within Kloudlite"
                disabled
                suffix={
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
                    <button
                      aria-label="copy"
                      onClick={() =>
                        copy(
                          `${consoleBaseUrl}/${parseName(account)}/${
                            project.name
                          }`
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
                value={project.name}
                label="Project ID"
                message="Used when interacting with the Kloudlite API"
                suffix={
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
                    <button
                      aria-label="copy"
                      onClick={() => copy(project.name)}
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
          title="Delete Project"
          action={() => {
            setDeleteProject(true);
          }}
        >
          Permanently remove your Project and all of its contents from the
          Kloudlite platform. This action is not reversible — please continue
          with caution.
        </DeleteContainer>
        <DeleteDialog
          resourceName={project.name}
          resourceType="project"
          show={deleteProject}
          setShow={setDeleteProject}
          onSubmit={async () => {
            try {
              const { errors } = await api.deleteIotProject({
                name: project.name,
              });

              if (errors) {
                throw errors[0];
              }
              reload();
              toast.success(`Project deleted successfully`);
              setDeleteProject(false);
              navigate(`/${parseName(account)}/environments`);
            } catch (err) {
              handleError(err);
            }
          }}
        />
      </Wrapper>
    </div>
  );
};
export default ProjectSettingGeneral;
