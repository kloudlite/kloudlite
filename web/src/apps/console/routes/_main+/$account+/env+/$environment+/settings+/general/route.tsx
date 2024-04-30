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
import useClipboard from '~/lib/client/hooks/use-clipboard';
import useForm, { dummyEvent } from '~/lib/client/hooks/use-form';
import { useUnsavedChanges } from '~/lib/client/hooks/use-unsaved-changes';
import { consoleBaseUrl } from '~/lib/configs/base-url.cjs';
import Yup from '~/lib/server/helpers/yup';
import { handleError } from '~/lib/utils/common';
import DeleteDialog from '~/console/components/delete-dialog';
import { useReload } from '~/lib/client/helpers/reloader';
import Wrapper from '~/console/components/wrapper';
import { Checkbox } from '~/components/atoms/checkbox';
import Banner from '~/components/molecule/banner';
import { IEnvironmentContext } from '../../_layout';

const EnvironmentSettingsGeneral = () => {
  const { environment, account } = useOutletContext<IEnvironmentContext>();

  const { setHasChanges, resetAndReload } = useUnsavedChanges();
  const [success, setSuccess] = useState(false);
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
      environmentRoutingMode: environment.spec?.routing?.mode === 'public',
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required('Name is required.'),
    }),
    onSubmit: async (val) => {
      try {
        const { errors: e } = await api.updateEnvironment({
          env: {
            displayName: val.displayName,
            metadata: {
              name: parseName(environment),
            },
            clusterName: environment.clusterName,
            spec: {
              routing: {
                mode: val.environmentRoutingMode ? 'public' : 'private',
              },
            },
          },
        });
        if (e) {
          throw e[0];
        }
        setSuccess(true);
      } catch (err) {
        handleError(err);
      }

      resetAndReload();
    },
  });

  const hasChanges = () =>
    values.displayName !== environment.displayName ||
    values.environmentRoutingMode !==
      (environment.spec?.routing?.mode === 'public');

  useEffect(() => {
    setHasChanges(hasChanges());
  }, [values]);

  useEffect(() => {
    resetValues();
  }, [environment]);

  const location = useLocation();

  useEffect(() => {
    setSuccess(false);
  }, [location]);

  return (
    <div>
      <Wrapper
        secondaryHeader={{
          title: 'General',
          action: hasChanges() && !success && (
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
        <Box title="General">
          <TextInput
            label="Environment name"
            value={values.displayName}
            onChange={handleChange('displayName')}
          />
          <Checkbox
            label="Public"
            checked={values.environmentRoutingMode}
            onChange={(checked) => {
              handleChange('environmentRoutingMode')(dummyEvent(checked));
            }}
          />
          <Banner
            type="info"
            body={
              <span>
                Public environments will expose services to the public internet.
                Private environments will be accessible when Kloudlite VPN is
                active.
              </span>
            }
          />
          <div className="flex flex-row items-center gap-3xl">
            <div className="flex-1">
              <TextInput
                label="Environment URL"
                value={`${consoleBaseUrl}/${parseName(account)}/${parseName(
                  environment
                )}`}
                message="This is your URL namespace within Kloudlite"
                disabled
                suffix={
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
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
                  <div
                    className="flex justify-center items-center"
                    title="Copy"
                  >
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
          title="Delete Environment"
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
              });

              if (errors) {
                throw errors[0];
              }
              reload();
              toast.success(`Environment deleted successfully`);
              setDeleteEnvironment(false);
              navigate(`/${parseName(account)}/environments/`);
            } catch (err) {
              handleError(err);
            }
          }}
        />
      </Wrapper>
    </div>
  );
};
export default EnvironmentSettingsGeneral;
