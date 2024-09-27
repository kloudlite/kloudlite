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
import { IDeviceBlueprints } from '~/iotconsole/server/gql/queries/iot-device-blueprint-queries';
import Select from '~/components/atoms/select';
import { IDeviceBlueprintContext } from '../../_layout';
import { deviceBlueprintTypes } from '../../../../deviceblueprints/blueprint-utils';

export const updateDeviceBlueprint = async ({
  api,
  data,
}: {
  api: ConsoleApiType;
  data: ExtractNodeType<IDeviceBlueprints>;
  projectName: string;
}) => {
  try {
    const { errors: e } = await api.updateIotDeviceBlueprint({
      deviceBlueprint: {
        name: data.name,
        displayName: data.displayName,
        bluePrintType: data.bluePrintType,
        version: data.version,
      },
    });
    if (e) {
      throw e[0];
    }
    toast.success('device blueprint updated successfully');
  } catch (err) {
    handleError(err);
  }
};

const ProjectSettingGeneral = () => {
  const { account, deviceblueprint } =
    useOutletContext<IDeviceBlueprintContext>();

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

  const { values, handleChange, submit, isLoading, resetValues } = useForm({
    initialValues: {
      displayName: deviceblueprint.displayName,
      deviceBlueprintType:
        deviceblueprint.bluePrintType || 'singleton_blueprint',
      version: deviceblueprint.version,
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required('Name is required.'),
    }),
    onSubmit: async (val) => {
      await updateDeviceBlueprint({
        api,
        data: {
          ...deviceblueprint,
          displayName: val.displayName,
          bluePrintType: val.deviceBlueprintType,
          version: val.version,
        },
        projectName: project.name,
      });
      resetAndReload();
    },
  });

  const checkChanges = () => {
    if (
      values.displayName !== deviceblueprint.displayName ||
      values.deviceBlueprintType !== deviceblueprint.bluePrintType ||
      values.version !== deviceblueprint.version
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
        <Box title="Device Blueprint details">
          <TextInput
            label="Device Blueprint name"
            value={values.displayName}
            onChange={handleChange('displayName')}
          />
          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <Select
                label="Blueprint type"
                value={values.deviceBlueprintType}
                options={async () => deviceBlueprintTypes}
                onChange={(_, value) => {
                  handleChange('deviceBlueprintType')(dummyEvent(value));
                }}
              />
            </div>
            <div className="flex-1">
              <TextInput
                label="Version"
                placeholder="Version"
                value={values.version}
                onChange={handleChange('version')}
              />
            </div>
          </div>
          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <TextInput
                label="Device Blueprint URL"
                value={`${consoleBaseUrl}/${parseName(account)}/${
                  project.name
                }/${deviceblueprint.name}`}
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
                          }/${deviceblueprint.name}`
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
                value={deviceblueprint.name}
                label="Device Blueprint ID"
                message="Used when interacting with the Kloudlite API"
                suffix={
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
                    <button
                      aria-label="copy"
                      onClick={() => copy(deviceblueprint.name)}
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
          title="Delete Device Blueprint"
          action={() => {
            setDeleteDeviceBlueprint(true);
          }}
        >
          Permanently remove your Device Blueprint and all of its contents from
          the Kloudlite platform. This action is not reversible — please
          continue with caution.
        </DeleteContainer>
        <DeleteDialog
          resourceName={deviceblueprint.name}
          resourceType="deviceblueprint"
          show={deleteDeviceBlueprint}
          setShow={setDeleteDeviceBlueprint}
          onSubmit={async () => {
            try {
              const { errors } = await api.deleteIotDeviceBlueprint({
                projectName: project.name,
                name: deviceblueprint.name,
              });

              if (errors) {
                throw errors[0];
              }
              reload();
              toast.success(`device blueprint deleted successfully`);
              setDeleteDeviceBlueprint(false);
              navigate(
                `/${parseName(account)}/${project.name}/deviceblueprints`
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
export default ProjectSettingGeneral;
