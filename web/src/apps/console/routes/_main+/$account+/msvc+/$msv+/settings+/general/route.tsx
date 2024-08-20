import { CopySimple } from '~/console/components/icons';
import { useLocation, useNavigate, useOutletContext } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/console/components/common-console-components';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName } from '~/console/server/r-utils/common';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import useForm from '~/root/lib/client/hooks/use-form';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import Wrapper from '~/console/components/wrapper';
import { useReload } from '~/root/lib/client/helpers/reloader';
import DeleteDialog from '~/console/components/delete-dialog';
import { getManagedTemplate } from '~/console/utils/commons';
import { IManagedServiceContext } from '../../_layout';
import { Fill } from '../../../../managed-services/handle-backend-service';

const ClusterManagedServiceSettingGeneral = () => {
  const { account, managedService, msvtemplates } =
    useOutletContext<IManagedServiceContext>();

  const { setHasChanges, resetAndReload } = useUnsavedChanges();
  const [success, setSuccess] = useState(false);
  const [deleteClusterMsvc, setDeleteClusterMsvc] = useState(false);

  const api = useConsoleApi();
  const reload = useReload();
  const navigate = useNavigate();

  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  const getService = () => {
    return getManagedTemplate({
      templates: msvtemplates,
      apiVersion:
        managedService.spec?.msvcSpec.serviceTemplate.apiVersion || '',
      kind: managedService.spec?.msvcSpec.serviceTemplate.kind || '',
    });
  };

  const { values, handleChange, submit, isLoading, resetValues, errors } =
    useForm({
      initialValues: {
        name: parseName(managedService),
        displayName: managedService.displayName,
        clusterName: managedService.clusterName,
        isNameError: false,
        res: {
          ...managedService.spec?.msvcSpec.serviceTemplate.spec,
        },
      },
      validationSchema: Yup.object({}),
      onSubmit: async (val) => {
        const { errors: e } = await api.updateClusterMSv({
          service: {
            displayName: val.displayName,
            metadata: {
              name: val.name,
            },
            clusterName: val.clusterName,
            spec: {
              msvcSpec: {
                serviceTemplate: {
                  apiVersion:
                    managedService.spec?.msvcSpec.serviceTemplate.apiVersion ||
                    '',
                  kind:
                    managedService.spec?.msvcSpec.serviceTemplate.kind || '',
                  spec: {
                    ...val.res,
                  },
                },
              },
            },
          },
        });
        if (e) {
          throw e[0];
        }
        toast.success('Integrated service updated successfully');
        setSuccess(true);
        resetAndReload();
      },
    });

  const checkChanges = () => {
    if (
      values.displayName !== managedService.displayName ||
      JSON.stringify(values.res) !==
        JSON.stringify(managedService.spec?.msvcSpec.serviceTemplate.spec)
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
  }, [managedService]);

  const location = useLocation();

  useEffect(() => {
    setSuccess(false);
  }, [location]);

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
        <Box title="Integrated service details">
          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <TextInput
                label="Integrated service URL"
                value={`${consoleBaseUrl}/${parseName(account)}/${parseName(
                  managedService,
                )}`}
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
                            managedService,
                          )}`,
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
                value={parseName(managedService)}
                label="Integrated service ID"
                message="Used when interacting with the Kloudlite API"
                suffix={
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
                    <button
                      aria-label="copy"
                      onClick={() => copy(parseName(managedService))}
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
          <Fill
            {...{
              selectedService: {
                category: { displayName: '', name: '' },
                service: getService(),
              },
              values,
              errors,
              handleChange,
            }}
            size="md"
          />
        </Box>

        <DeleteContainer
          title="Delete Integrated Service"
          action={() => {
            setDeleteClusterMsvc(true);
          }}
        >
          Permanently remove your Integrated service and all of its contents
          from the Kloudlite platform. This action is not reversible — please
          continue with caution.
        </DeleteContainer>
        <DeleteDialog
          resourceName={parseName(managedService)}
          resourceType="Integrated service"
          show={deleteClusterMsvc}
          setShow={setDeleteClusterMsvc}
          onSubmit={async () => {
            try {
              const { errors } = await api.deleteClusterMSv({
                name: parseName(managedService),
              });

              if (errors) {
                throw errors[0];
              }
              reload();
              toast.success(`Integrated service deleted successfully`);
              setDeleteClusterMsvc(false);
              navigate(`/${parseName(account)}/managed-services`);
            } catch (err) {
              handleError(err);
            }
          }}
        />
      </Wrapper>
    </div>
  );
};
export default ClusterManagedServiceSettingGeneral;
