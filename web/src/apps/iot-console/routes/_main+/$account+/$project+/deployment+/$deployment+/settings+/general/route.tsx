import { CopySimple } from '~/iotconsole/components/icons';
import { useNavigate, useOutletContext } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/iotconsole/components/common-console-components';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import { ConsoleApiType } from '~/iotconsole/server/gql/saved-queries';
import { ExtractNodeType, parseName } from '~/iotconsole/server/r-utils/common';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import Wrapper from '~/iotconsole/components/wrapper';
import { useReload } from '~/root/lib/client/helpers/reloader';
import DeleteDialog from '~/iotconsole/components/delete-dialog';
import Select from '~/components/atoms/select';
import { IDeployments } from '~/iotconsole/server/gql/queries/iot-deployment-queries';
import KeyValuePair from '~/iotconsole/components/key-value-pair';
import { IDeploymentContext } from '../../_layout';

export const updateDeployment = async ({
  api,
  data,
  
}: {
  api: ConsoleApiType;
  data: ExtractNodeType<IDeployments>;
  projectName: string;
}) => {
  try {
    const { errors: e } = await api.updateIotDeployment({
      
      deployment: {
        name: data.name,
        displayName: data.displayName,
        CIDR: data.CIDR,
        exposedIps: data.exposedIps,
        exposedDomains: data.exposedDomains,
        exposedServices: data.exposedServices.map((service) => {
          return {
            name: service.name,
            ip: service.ip,
          };
        }),
      },
    });
    if (e) {
      throw e[0];
    }
    toast.success('Deployment updated successfully');
  } catch (err) {
    handleError(err);
  }
};

const ProjectSettingGeneral = () => {
  const {  account, deployment } =
    useOutletContext<IDeploymentContext>();

  const { setHasChanges, resetAndReload } = useUnsavedChanges();

  const [deleteDeviceBlueprint, setDeleteDeviceBlueprint] = useState(false);

  const api = useIotConsoleApi();
  const reload = useReload();
  const navigate = useNavigate();

  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  const { values, handleChange, submit, isLoading, resetValues, errors } =
    useForm({
      initialValues: {
        displayName: deployment.displayName,
        cidr: deployment.CIDR,
        exposedServices: deployment.exposedServices,
        exposedIps: deployment.exposedIps,
        exposedDomains: deployment.exposedDomains,
      },
      validationSchema: Yup.object({
        displayName: Yup.string().required('Name is required.'),
      }),
      onSubmit: async (val) => {
        await updateDeployment({
          api,
          data: {
            ...deployment,
            displayName: val.displayName,
            CIDR: val.cidr,
            exposedServices: val.exposedServices,
            exposedIps: val.exposedIps,
            exposedDomains: val.exposedDomains,
          },
          projectName: project.name,
        });
        resetAndReload();
      },
    });

  const checkChanges = () => {
    if (
      values.displayName !== deployment.displayName ||
      values.exposedIps !== deployment.exposedIps ||
      values.exposedDomains !== deployment.exposedDomains
    ) {
      return true;
    }
    return false;
  };

  useEffect(() => {
    setHasChanges(checkChanges());
  }, [values]);

  useEffect(() => {
    resetValues();
  }, [project]);

  return (
    <div>
      <Wrapper
        secondaryHeader={{
          title: 'General',
          action: checkChanges() && (
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
        <Box title="Deployment details">
          <TextInput
            label="Deployment name"
            value={values.displayName}
            onChange={handleChange('displayName')}
          />

          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <TextInput
                label="Deployment URL"
                value={`${consoleBaseUrl}/${parseName(account)}/${
                  project.name
                }/${deployment.name}`}
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
                          }/${deployment.name}`
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
                value={deployment.name}
                label="Deployment ID"
                message="Used when interacting with the Kloudlite API"
                suffix={
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
                    <button
                      aria-label="copy"
                      onClick={() => copy(deployment.name)}
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

        <Box title="">
          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <Select
                creatable
                size="lg"
                label="Exposed ips"
                multiple
                value={values.exposedIps || []}
                options={async () => []}
                onChange={(val, v) => {
                  handleChange('exposedIps')(dummyEvent(v));
                }}
                // error={!!errors.exposedIps}
                disableWhileLoading
              />
            </div>
            <div className="flex-1">
              <Select
                creatable
                size="lg"
                label="Exposed domains"
                multiple
                value={values.exposedDomains || []}
                options={async () => []}
                onChange={(val, v) => {
                  handleChange('exposedDomains')(dummyEvent(v));
                }}
                // error={!!errors.exposedDomains}
                disableWhileLoading
              />
            </div>
          </div>
          <KeyValuePair
            size="lg"
            label="Exposed services"
            value={values.exposedServices || []}
            onChange={(items, _) => {
              handleChange('exposedServices')(dummyEvent(items));
            }}
            keyLabel="name"
            valueLabel="ip"
            error={!!errors.exposedServices}
            message={errors.exposedServices}
          />
        </Box>

        <DeleteContainer
          title="Delete Deployment"
          action={() => {
            setDeleteDeviceBlueprint(true);
          }}
        >
          Permanently remove your Deployment and all of its contents from the
          Kloudlite platform. This action is not reversible — please continue
          with caution.
        </DeleteContainer>
        <DeleteDialog
          resourceName={deployment.name}
          resourceType="deviceblueprint"
          show={deleteDeviceBlueprint}
          setShow={setDeleteDeviceBlueprint}
          onSubmit={async () => {
            try {
              const { errors } = await api.deleteIotDeployment({
                projectName: project.name,
                name: deployment.name,
              });

              if (errors) {
                throw errors[0];
              }
              reload();
              toast.success(`Deployment deleted successfully`);
              setDeleteDeviceBlueprint(false);
              navigate(`/${parseName(account)}/${project.name}/deployments`);
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
