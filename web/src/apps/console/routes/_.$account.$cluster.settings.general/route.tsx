/* eslint-disable jsx-a11y/control-has-associated-label */
import { CopySimple } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData, useOutletContext } from '@remix-run/react';
import { useEffect } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { toast } from '~/components/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/console/components/common-console-components';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import SubNavAction from '~/console/components/sub-nav-action';
import { awsRegions } from '~/console/dummy/consts';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ICluster } from '~/console/server/gql/queries/cluster-queries';
import {
  ConsoleApiType,
  GQLServerHandler,
} from '~/console/server/gql/saved-queries';
import {
  ensureResource,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination } from '~/console/server/utils/common';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import useForm from '~/root/lib/client/hooks/use-form';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import Yup from '~/root/lib/server/helpers/yup';
import { IRemixCtx } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import { mapper } from '~/components/utils';
import { IClusterContext } from '../_.$account.$cluster';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listProviderSecrets({
      pagination: getPagination(ctx),
    });

    if (errors) {
      throw errors[0];
    }

    return {
      providerSecrets: data,
    };
  });
  return defer({ promise });
};

export const updateCluster = async ({
  api,
  data,
}: {
  api: ConsoleApiType;
  data: ICluster;
}) => {
  try {
    const { errors: e } = await api.updateCluster({
      cluster: {
        displayName: data.displayName,
        metadata: {
          name: data.metadata.name,
        },
        spec: ensureResource(data.spec),
      },
    });
    if (e) {
      throw e[0];
    }
  } catch (err) {
    handleError(err);
  }
};

const SettingGeneral = () => {
  const { promise } = useLoaderData<typeof loader>();

  const { account, cluster } = useOutletContext<IClusterContext>();

  const { setHasChanges, resetAndReload } = useUnsavedChanges();
  const api = useConsoleApi();

  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  const { values, handleChange, submit, isLoading, resetValues, setValues } =
    useForm({
      initialValues: {
        displayName: account.displayName,
      },
      validationSchema: Yup.object({
        displayName: Yup.string().required('Name is required.'),
      }),
      onSubmit: async (val) => {
        await updateCluster({
          api,
          data: { ...cluster, displayName: val.displayName },
        });
        resetAndReload();
      },
    });

  useEffect(() => {
    setValues({
      displayName: account.displayName,
    });
  }, [account]);

  useEffect(() => {
    setHasChanges(values.displayName !== account.displayName);
  }, [values]);

  return (
    <LoadingComp data={promise}>
      {({ providerSecrets }) => {
        const providerSecretsOptions = parseNodes(providerSecrets).map(
          (provider) => ({
            value: parseName(provider),
            label: provider.displayName,
            render: () => (
              <div className="flex flex-col">
                <div>{provider.displayName}</div>
                <div className="bodySm text-text-soft">
                  {parseName(provider)}
                </div>
              </div>
            ),
            provider,
          })
        );

        const defaultProvider = providerSecretsOptions.find(
          (ps) => ps.value === cluster.spec?.credentialsRef.name
        );

        const defaultRegion = awsRegions.find(
          (r) => r.Name === cluster.spec?.aws?.region
        );

        return (
          <div className="flex flex-col gap-6xl">
            <SubNavAction deps={[values, isLoading]}>
              {values.displayName !== account.displayName && (
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
                    onClick={() => {
                      submit();
                    }}
                    loading={isLoading}
                  />
                </>
              )}
            </SubNavAction>
            <Box title="General">
              <div className="flex flex-row items-center gap-3xl">
                <div className="flex-1">
                  <TextInput
                    label="Cluster name"
                    value={values.displayName}
                    onChange={handleChange('displayName')}
                  />
                </div>
                <div className="flex-1">
                  <TextInput
                    value={cluster.metadata.name}
                    label="Cluster ID"
                    suffix={
                      <div
                        className="flex justify-center items-center"
                        title="Copy"
                      >
                        <button
                          onClick={() => copy(cluster.metadata.name)}
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
              <div className="flex flex-row items-center gap-3xl">
                <div className="flex-1">
                  {' '}
                  <Select
                    label="Cloud Provider"
                    placeholder="Select cloud provider"
                    disabled
                    value={defaultProvider}
                    options={async () => providerSecretsOptions}
                  />
                </div>
                <div className="flex-1">
                  <Select
                    disabled
                    label="Region"
                    placeholder="Select region"
                    value={{
                      value: defaultRegion?.Name || '',
                      label: defaultRegion?.Name || '',
                      region: defaultRegion as any,
                    }}
                    options={async () =>
                      mapper(awsRegions, (v) => {
                        return {
                          value: v.Name,
                          label: v.Name,
                          region: v,
                        };
                      })
                    }
                  />
                </div>
              </div>
            </Box>

            <DeleteContainer
              title="Delete Cluster"
              action={async () => {
                await api.deleteCluster({
                  name: cluster.metadata.name,
                });
              }}
            >
              Permanently remove your Cluster and all of its contents from the
              Kloudlite platform. This action is not reversible — please
              continue with caution.
            </DeleteContainer>
          </div>
        );
      }}
    </LoadingComp>
  );
};
export default SettingGeneral;
