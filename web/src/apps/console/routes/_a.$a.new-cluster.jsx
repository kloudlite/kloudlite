import { Button } from '~/components/atoms/button';
import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import { useState } from 'react';
import {
  useParams,
  useLoaderData,
  useOutletContext,
  useNavigate,
} from '@remix-run/react';
import * as SelectInput from '~/components/atoms/select';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import * as Tooltip from '~/components/atoms/tooltip';
import logger from '~/root/lib/client/helpers/log';
import { toast } from '~/components/molecule/toast';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { Select } from '~/components/atoms/select-new';
import { IdSelector, idTypes } from '../components/id-selector';
import {
  getCluster,
  getClusterSepc,
  getCredentialsRef,
} from '../server/r-urils/cluster';
import {
  getMetadata,
  getPagination,
  parseName,
} from '../server/r-urils/common';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { keyconstants } from '../server/r-urils/key-constants';
import { constDatas } from '../dummy/consts';
import { ensureAccountSet } from '../server/utils/auth-utils';
import AlertDialog from '../components/alert-dialog';

const NewCluster = () => {
  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);
  const api = useAPIClient();

  const { providerSecrets } = useLoaderData();
  const cloudProviders = providerSecrets?.edges?.map(({ node }) => node) || [];

  const { a: account } = useParams();
  // @ts-ignore
  const { user } = useOutletContext();

  const navigate = useNavigate();

  const [selectedProvider, setSelectedProvider] = useState();

  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      vpc: '',
      name: '',
      region: 'ap-south-1',
      cloudProvider: '',
      credentialsRef: '',
      availabilityMode: '',
      displayName: '',
    },
    validationSchema: Yup.object({
      vpc: Yup.string(),
      region: Yup.string().trim().required('region is required'),
      cloudProvider: Yup.string().trim().required('cloud provider is required'),
      name: Yup.string().trim().required('id is required'),
      displayName: Yup.string().trim().required('name is required'),
      credentialsRef: Yup.string().required(),
      availabilityMode: Yup.string()
        .trim()
        .required('availability is required'),
    }),
    onSubmit: async (val) => {
      try {
        const { errors: e } = await api.createCluster({
          cluster: getCluster({
            spec: getClusterSepc({
              accountName: account,
              vpc: val.vpc || undefined,
              region: val.region,
              cloudProvider: val.cloudProvider,
              credentialsRef: getCredentialsRef({
                name: val.credentialsRef,
              }),
              availabilityMode: val.availabilityMode,
            }),
            metadata: getMetadata({
              name: val.name,
              annotations: {
                [keyconstants.displayName]: val.displayName,
                [keyconstants.author]: user.name,
              },
            }),
          }),
        });
        if (e) {
          throw e[0];
        }
        toast.success('cluster created successfully');
        navigate(`/${account}/clusters`);
      } catch (err) {
        toast.error(err.message);
      }
    },
  });

  return (
    <Tooltip.TooltipProvider>
      <div className="h-full flex flex-row">
        <div className="h-full w-[571px] flex flex-col bg-surface-basic-subdued py-11xl px-10xl">
          <div className="flex flex-col gap-8xl">
            <div className="flex flex-col gap-4xl items-start">
              <BrandLogo detailed={false} size={48} />
              <div className="flex flex-col gap-3xl">
                <div className="text-text-default heading4xl">
                  Let’s create new cluster.
                </div>
                <div className="text-text-default bodyLg">
                  Create your cluster to production effortlessly
                </div>
              </div>
            </div>
            <ProgressTracker
              items={[
                { label: 'Configure cluster', active: true, id: 1 },
                { label: 'Review', active: false, id: 2 },
              ]}
            />
            <Button
              variant="outline"
              content="Back"
              prefix={ArrowLeft}
              onClick={() => setShowUnsavedChanges(true)}
            />
          </div>
        </div>
        <form className="py-11xl px-10xl flex-1" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4xl">
            <div className="h-7xl" />
            <div className="flex flex-col gap-3xl p-3xl">
              {Object.keys(JSON.parse(JSON.stringify(errors || '{}')) || {})
                .length > 0 && (
                <pre className="text-xs text-surface-warning-default">
                  <code>{JSON.stringify(errors, null, 2)}</code>
                </pre>
              )}
              <TextInput
                label="Cluster name"
                onChange={handleChange('displayName')}
                value={values.displayName}
                error={!!errors.displayName}
                message={errors.displayName}
              />
              <IdSelector
                resType={idTypes.cluster}
                name={values.displayName}
                onChange={(v) => {
                  handleChange('name')({ target: { value: v } });
                }}
              />

              <Select
                id="cloudprovider-select"
                label="Cloud Provider"
                value={{
                  value: parseName(selectedProvider),
                  label: parseName(selectedProvider),
                  provider: selectedProvider,
                }}
                options={(cloudProviders || []).map((provider) => ({
                  value: parseName(provider),
                  label: parseName(provider),
                  provider,
                }))}
                onChange={({ provider }) => {
                  handleChange('credentialsRef')({
                    target: { value: parseName(provider) },
                  });
                  handleChange('cloudProvider')({
                    target: { value: provider?.cloudProviderName },
                  });
                  setSelectedProvider(provider);
                }}
              />

              <SelectInput.Select
                label="Region"
                value={values.region}
                onChange={(v) => {
                  handleChange('region')({ target: { value: v } });
                }}
              >
                <SelectInput.Option> -- not-selected -- </SelectInput.Option>
                {constDatas.regions.map(({ name, value }) => {
                  return (
                    <SelectInput.Option key={value} value={value}>
                      {name}
                    </SelectInput.Option>
                  );
                })}
              </SelectInput.Select>

              <SelectInput.Select
                label="Availabilty"
                value={values.availabilityMode}
                onChange={(v) => {
                  handleChange('availabilityMode')({ target: { value: v } });
                }}
              >
                <SelectInput.Option> -- not-selected -- </SelectInput.Option>
                {constDatas.availabilityModes.map(({ name, value }) => {
                  return (
                    <SelectInput.Option key={value} value={value}>
                      {name}
                    </SelectInput.Option>
                  );
                })}
              </SelectInput.Select>

              <TextInput
                label="VPC"
                onChange={handleChange('vpc')}
                value={values.vpc}
                error={!!errors.vpc}
                message={errors.vpc}
              />
            </div>
          </div>
          <div className="flex flex-row justify-end px-3xl">
            <Button
              loading={isLoading}
              variant="primary"
              content="Create"
              suffix={ArrowRight}
              type="submit"
            />
          </div>
        </form>

        {/* Unsaved change alert dialog */}

        <AlertDialog
          title="Leave page with unsaved changes?"
          message="Leaving this page will delete all unsaved changes."
          okText="Delete"
          type="critical"
          show={showUnsavedChanges}
          setShow={setShowUnsavedChanges}
          onSubmit={() => {
            setShowUnsavedChanges(false);
            navigate(`/${account}/clusters`);
          }}
        />
      </div>
    </Tooltip.TooltipProvider>
  );
};

export const loader = async (ctx) => {
  ensureAccountSet(ctx);
  const { data, errors } = await GQLServerHandler(
    ctx.request
  ).listProviderSecrets({
    pagination: getPagination(ctx),
  });

  if (errors) {
    logger.error(errors);
  }

  return {
    providerSecrets: data,
  };
};

export default NewCluster;
