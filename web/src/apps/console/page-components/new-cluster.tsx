import { ArrowLeft, ArrowRight, UserCircle } from '@jengaicons/react';
import { useNavigate, useParams } from '@remix-run/react';
import { useMemo, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { toast } from '~/components/molecule/toast';
import { mapper, useMapper } from '~/components/utils';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import AlertModal from '../components/alert-modal';
import { IdSelector } from '../components/id-selector';
import RawWrapper, { TitleBox } from '../components/raw-wrapper';
import { constDatas, regions } from '../dummy/consts';
import { FadeIn } from '../routes/_.$account.$cluster.$project.$scope.$workspace.new-app/util';
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
  validateClusterCloudProvider,
} from '../server/r-utils/common';
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
    () => parseNodes(providerSecrets!),
    [providerSecrets]
  );

  const options = useMapper(cloudProviders, (provider) => ({
    value: parseName(provider),
    label: provider.displayName,
    provider,
    render: () => (
      <div className="flex flex-col">
        <div>{provider.displayName}</div>
        <div className="bodySm text-text-soft">{parseName(provider)}</div>
      </div>
    ),
  }));

  const { a: accountName } = useParams();

  const navigate = useNavigate();

  const [selectedProvider, setSelectedProvider] = useState<
    | {
        label: string;
        value: string;
        provider: ExtractNodeType<IProviderSecrets>;
        render: () => JSX.Element;
      }
    | undefined
  >(options.length === 1 ? options[0] : undefined);

  const [selectedRegion, setSelectedRegion] = useState<
    (typeof regions)[number]
  >(regions[0]);

  const [selectedAvailabilityMode, setSelectedAvailabilityMode] = useState<
    (typeof constDatas.availabilityModes)[number] | undefined
  >();

  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      vpc: '',
      name: '',
      region: 'ap-south-1' || selectedRegion?.Name,
      cloudProvider: cloudProvider
        ? cloudProvider.cloudProviderName
        : selectedProvider?.provider?.cloudProviderName || '',
      credentialsRef: cp || parseName(selectedProvider?.provider) || '',
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
              cloudProvider: validateClusterCloudProvider(val.cloudProvider),
              aws: {
                region: selectedRegion.Name,
                k3sMasters: {
                  instanceType: 'c6a.xlarge',
                },
              },
              credentialsRef: {
                name: val.credentialsRef,
              },
              availabilityMode: validateAvailabilityMode(val.availabilityMode),
            },
            metadata: {
              name: val.name,
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
        badge={{
          title: 'Kloudlite Labs Pvt Ltd',
          subtitle: accountName,
          image: <UserCircle size={20} />,
        }}
        onCancel={() => setShowUnsavedChanges(true)}
        rightChildren={
          <FadeIn onSubmit={handleSubmit}>
            <TitleBox
              title="Cluster details"
              subtitle="A cluster is a group of interconnected elements working together
                as a single unit."
            />
            <div className="flex flex-col">
              <div className="flex flex-col gap-3xl pb-xl">
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
              </div>
              <IdSelector
                resType="cluster"
                name={values.displayName}
                onChange={(v) => {
                  handleChange('name')({ target: { value: v } });
                }}
              />
              <div className="flex flex-col gap-3xl pt-lg">
                {!isOnboarding && (
                  <Select
                    label="Cloud Provider"
                    size="lg"
                    placeholder="Select cloud provider"
                    value={selectedProvider}
                    options={async () => options}
                    onChange={(value) => {
                      handleChange('credentialsRef')({
                        target: { value: parseName(value.provider) },
                      });
                      handleChange('cloudProvider')({
                        target: {
                          value: value.provider?.cloudProviderName || '',
                        },
                      });
                      setSelectedProvider(value);
                    }}
                  />
                )}
                <Select
                  label="Region"
                  size="lg"
                  placeholder="Select region"
                  value={{
                    label: selectedRegion?.Name || '',
                    value: selectedRegion?.Name || '',
                    region: selectedRegion,
                  }}
                  options={async () =>
                    mapper(regions, (v) => {
                      return {
                        value: v.Name,
                        label: v.Name,
                        region: v,
                      };
                    })
                  }
                  onChange={(region) => {
                    handleChange('region')(dummyEvent(region.value));
                    setSelectedRegion(region.region);
                  }}
                />

                <Select
                  label="Availabity"
                  size="lg"
                  placeholder="Select availability mode"
                  value={selectedAvailabilityMode}
                  options={async () => constDatas.availabilityModes}
                  onChange={(availabilityMode) => {
                    handleChange('availabilityMode')(
                      dummyEvent(availabilityMode.value)
                    );
                    setSelectedAvailabilityMode(availabilityMode);
                  }}
                />

                <TextInput
                  label="VPC"
                  size="lg"
                  onChange={handleChange('vpc')}
                  value={values.vpc}
                  error={!!errors.vpc}
                  message={errors.vpc}
                />
              </div>
            </div>
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
          </FadeIn>
        }
      />
      <AlertModal
        title="Leave page with unsaved changes?"
        message="Leaving this page will delete all unsaved changes."
        okText="Leave"
        cancelText="Stay"
        variant="critical"
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
