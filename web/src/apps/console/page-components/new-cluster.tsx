import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { useNavigate, useOutletContext, useParams } from '@remix-run/react';
import { useMemo, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import SelectInput from '~/components/atoms/select-primitive';
import { toast } from '~/components/molecule/toast';
import { useMapper } from '~/components/utils';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import AlertDialog from '../components/alert-dialog';
import { IdSelector } from '../components/id-selector';
import RawWrapper from '../components/raw-wrapper';
import { constDatas } from '../dummy/consts';
import { useConsoleApi } from '../server/gql/api-provider';
import {
  IProviderSecret,
  IProviderSecrets,
} from '../server/gql/queries/provider-secret-queries';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
  validateAvailabilityMode,
  validateCloudProvider,
} from '../server/r-utils/common';
import { keyconstants } from '../server/r-utils/key-constants';
import { ensureAccountClientSide } from '../server/utils/auth-utils';

type props =
  | {
      providerSecrets: IProviderSecrets;
      cloudProvider?: IProviderSecret;
    }
  | {
      providerSecrets?: IProviderSecrets;
      cloudProvider: IProviderSecret;
    };

export const NewCluster = ({ providerSecrets, cloudProvider }: props) => {
  const { cloudprovider: cp } = useParams();
  const isOnboarding = !!cp;

  const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);
  const api = useConsoleApi();

  const cloudProviders = useMemo(
    () => parseNodes(providerSecrets),
    [providerSecrets]
  );

  const { a: accountName } = useParams();
  const { user } = useOutletContext<{
    user: any;
    account: any;
  }>();

  const navigate = useNavigate();
  const k: any = null;
  const [selectedProvider, setSelectedProvider] =
    useState<ExtractNodeType<IProviderSecrets>>(k);

  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      vpc: '',
      name: '',
      region: 'ap-south-1',
      cloudProvider: cloudProvider ? cloudProvider.cloudProviderName : '',
      credentialsRef: cp || '',
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
        .required('availability is required')
        .oneOf(['HA', 'dev']),
    }),
    onSubmit: async (val) => {
      type Merge<T, M> = Omit<T, keyof M> & M;

      type nt = { availabilityMode: 'HA' | 'dev' | string };
      const k: Merge<typeof val, nt> = val;

      console.log(k);
      // val.availabilityMode
      if (!accountName || !val.availabilityMode) {
        return;
      }
      try {
        ensureAccountClientSide({ account: accountName });
        const { errors: e } = await api.createCluster({
          cluster: {
            displayName: val.displayName,
            spec: {
              accountName,
              vpc: val.vpc || undefined,
              region: val.region,
              cloudProvider: validateCloudProvider(val.cloudProvider),
              credentialsRef: {
                name: val.credentialsRef,
              },
              availabilityMode: validateAvailabilityMode(val.availabilityMode),
            },
            metadata: {
              name: val.name,
              annotations: {
                [keyconstants.author]: user.name,
              },
            },
          },
        });
        if (e) {
          throw e[0];
        }
        toast.success('cluster created successfully');
        navigate(
          isOnboarding
            ? `/onboarding/${accountName}/${val.name}/new-project`
            : `/${accountName}/clusters`
        );
      } catch (err) {
        handleError(err);
      }
    },
  });

  const items = useMapper(
    isOnboarding
      ? [
          {
            label: 'Create Team',
            active: true,
            id: 1,
            completed: false,
          },
          {
            label: 'Invite your Team Members',
            active: true,
            id: 2,
            completed: false,
          },
          {
            label: 'Add your Cloud Provider',
            active: true,
            id: 3,
            completed: false,
          },
          {
            label: 'Setup First Cluster',
            active: true,
            id: 4,
            completed: false,
          },
          {
            label: 'Create your project',
            active: false,
            id: 5,
            completed: false,
          },
        ]
      : [
          {
            label: 'Configure cluster',
            active: true,
            id: 1,
            completed: false,
          },
          {
            label: 'Review',
            active: false,
            id: 2,
            completed: false,
          },
        ],
    (i) => {
      return {
        value: i.id,
        item: {
          ...i,
        },
      };
    }
  );

  const options = useMapper(cloudProviders, (provider) => ({
    value: parseName(provider),
    label: parseName(provider),
    provider,
  }));
  // useLog(options);
  return (
    <>
      <RawWrapper
        title={
          isOnboarding
            ? "Unleash Data's Full Potential!"
            : 'Let’s create new cluster.'
        }
        subtitle={
          isOnboarding
            ? 'Kloudlite will help you to develop and deploy cloud native applications easily.'
            : 'Create your cluster under to production effortlessly'
        }
        progressItems={items}
        rightChildren={
          <form onSubmit={handleSubmit} className="flex flex-col gap-3xl">
            <div className="text-text-soft headingLg">Cluster details</div>
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
              size="lg"
            />
            <IdSelector
              resType="cluster"
              name={values.displayName}
              onChange={(v) => {
                handleChange('name')({ target: { value: v } });
              }}
            />

            {!isOnboarding && (
              <Select
                label="Cloud Provider"
                size="lg"
                placeholder="--- Select ---"
                value={{
                  label: parseName(selectedProvider),
                  value: parseName(selectedProvider),
                  provider: selectedProvider,
                }}
                // multiselect
                options={options}
                onChange={({ provider }) => {
                  handleChange('credentialsRef')({
                    target: { value: parseName(provider) },
                  });
                  handleChange('cloudProvider')({
                    target: { value: provider?.cloudProviderName || '' },
                  });
                  setSelectedProvider(provider);
                }}
              />
            )}

            <SelectInput.Root
              label="Region"
              value={values.region}
              size="lg"
              onChange={handleChange('region')}
            >
              <SelectInput.Option> -- not-selected -- </SelectInput.Option>
              {constDatas.regions.map(({ name, value }) => {
                return (
                  <SelectInput.Option key={value} value={value}>
                    {name}
                  </SelectInput.Option>
                );
              })}
            </SelectInput.Root>

            <SelectInput.Root
              label="Availabilty"
              size="lg"
              value={values.availabilityMode}
              onChange={handleChange('availabilityMode')}
            >
              <SelectInput.Option> -- not-selected -- </SelectInput.Option>
              {constDatas.availabilityModes.map(({ name, value }) => {
                return (
                  <SelectInput.Option key={value} value={value}>
                    {name}
                  </SelectInput.Option>
                );
              })}
            </SelectInput.Root>

            <TextInput
              label="VPC"
              size="lg"
              onChange={handleChange('vpc')}
              value={values.vpc}
              error={!!errors.vpc}
              message={errors.vpc}
            />
            {isOnboarding ? (
              <div className="flex flex-row gap-xl justify-end">
                <Button
                  variant="outline"
                  content="Back"
                  prefix={<ArrowLeft />}
                  size="lg"
                />
                <Button
                  variant="primary"
                  content="Continue"
                  suffix={<ArrowRight />}
                  size="lg"
                  type="submit"
                />
              </div>
            ) : (
              <div className="flex flex-row justify-end">
                <Button
                  loading={isLoading}
                  variant="primary"
                  content="Create"
                  suffix={<ArrowRight />}
                  type="submit"
                  size="lg"
                />
              </div>
            )}
          </form>
        }
      />
      <AlertDialog
        title="Leave page with unsaved changes?"
        message="Leaving this page will delete all unsaved changes."
        okText="Delete"
        type="critical"
        show={showUnsavedChanges}
        setShow={setShowUnsavedChanges}
        onSubmit={() => {
          setShowUnsavedChanges(false);
          navigate(`/${accountName}/clusters`);
        }}
      />
    </>
  );
};
