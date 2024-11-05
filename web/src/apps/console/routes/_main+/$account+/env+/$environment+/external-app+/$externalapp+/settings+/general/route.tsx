import { CopySimple } from '~/console/components/icons';
import { useNavigate, useOutletContext } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button } from '@kloudlite/design-system/atoms/button';
import { TextInput } from '@kloudlite/design-system/atoms/input';
import { toast } from '@kloudlite/design-system/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/console/components/common-console-components';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ConsoleApiType } from '~/console/server/gql/saved-queries';
import {
  ExtractNodeType,
  parseName,
  validateExternalAppRecordType,
} from '~/console/server/r-utils/common';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import useForm from '~/root/lib/client/hooks/use-form';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import Wrapper from '~/console/components/wrapper';
import { useReload } from '~/root/lib/client/helpers/reloader';
import DeleteDialog from '~/console/components/delete-dialog';
import { IExternalApps } from '~/console/server/gql/queries/external-app-queries';
import { IExternalAppContext } from '../../_layout';

export const updateApp = async ({
  api,
  data,
  environmentName,
}: {
  api: ConsoleApiType;
  data: ExtractNodeType<IExternalApps>;
  environmentName: string;
}) => {
  try {
    const { errors: e } = await api.updateExternalApp({
      externalApp: {
        displayName: data.displayName,
        metadata: {
          name: parseName(data),
        },
        spec: {
          record: data.spec?.record || '',
          recordType: validateExternalAppRecordType(
            data.spec?.recordType || ''
          ),
        },
      },
      envName: environmentName,
    });
    if (e) {
      throw e[0];
    }
    toast.success('External App updated successfully');
  } catch (err) {
    handleError(err);
  }
};

const ProjectSettingGeneral = () => {
  const { account, environment, app } = useOutletContext<IExternalAppContext>();

  const { setHasChanges, resetAndReload } = useUnsavedChanges();
  const [success, setSuccess] = useState(false);

  const [deleteDeviceBlueprint, setDeleteDeviceBlueprint] = useState(false);

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
      displayName: app.displayName,
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required('Name is required.'),
    }),
    onSubmit: async (val) => {
      await updateApp({
        api,
        data: {
          ...app,
          displayName: val.displayName,
        },
        environmentName: parseName(environment),
      });
      setSuccess(true);
      resetAndReload();
    },
  });

  const checkChanges = () => {
    if (
      values.displayName !== app.displayName
      //   values.exposedIps !== deployment.exposedIps ||
      //   values.exposedDomains !== deployment.exposedDomains
    ) {
      return true;
    }
    return false;
  };

  useEffect(() => {
    setHasChanges(checkChanges());
  }, [values]);

  return (
    <div>
      <Wrapper
        secondaryHeader={{
          title: 'General',
          action: checkChanges() && !success && (
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
        <Box title="External App details">
          <TextInput
            label="External app name"
            value={values.displayName}
            onChange={handleChange('displayName')}
          />

          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <TextInput
                label="External app URL"
                value={`${consoleBaseUrl}/${parseName(account)}/${parseName(
                  environment
                )}/${parseName(app)}`}
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
                          `${consoleBaseUrl}/${parseName(account)}/${parseName(
                            environment
                          )}/${parseName(app)}`
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
                value={parseName(app)}
                label="Deployment ID"
                message="Used when interacting with the Kloudlite API"
                suffix={
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
                    <button
                      aria-label="copy"
                      onClick={() => copy(parseName(app))}
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

        {/* <Box title="">
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
        </Box> */}

        <DeleteContainer
          title="Delete External App"
          action={() => {
            setDeleteDeviceBlueprint(true);
          }}
        >
          Permanently remove your Deployment and all of its contents from the
          Kloudlite platform. This action is not reversible — please continue
          with caution.
        </DeleteContainer>
        <DeleteDialog
          resourceName={parseName(app)}
          resourceType="external app"
          show={deleteDeviceBlueprint}
          setShow={setDeleteDeviceBlueprint}
          onSubmit={async () => {
            try {
              const { errors } = await api.deleteExternalApp({
                envName: parseName(environment),
                externalAppName: parseName(app),
              });

              if (errors) {
                throw errors[0];
              }
              reload();
              toast.success(`external app deleted successfully`);
              setDeleteDeviceBlueprint(false);
              navigate(
                `/${parseName(account)}/env/${parseName(
                  environment
                )}/external-apps`
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
