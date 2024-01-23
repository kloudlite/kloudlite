import { ArrowRight } from '@jengaicons/react';
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
import { constDatas, awsRegions } from '../dummy/consts';
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
import { NameIdView } from '../components/name-id-view';
import { ReviewComponent } from '../routes/_main+/$account+/$project+/$environment+/new-app/app-review';
import MultiStepProgress, {
  useMultiStepProgress,
} from '../components/multi-step-progress';
import MultiStepProgressWrapper from '../components/multi-step-progress-wrapper';
import { TitleBox } from '../components/raw-wrapper';

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

  const { currentStep, jumpStep, nextStep } = useMultiStepProgress({
    defaultStep: isOnboarding ? 4 : 1,
    totalSteps: isOnboarding ? 4 : 2,
  });

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
    (typeof awsRegions)[number]
  >(awsRegions[0]);

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
      isNameError: false,
    },
    validationSchema: Yup.object({
      vpc: Yup.string(),
      region: Yup.string().trim().required('Region is required'),
      cloudProvider: Yup.string().trim().required('Cloud provider is required'),
      name: Yup.string().trim().required('Name is required'),
      displayName: Yup.string().trim().required('Name is required'),
      credentialsRef: Yup.string().required(),
      availabilityMode: Yup.string()
        .trim()
        .oneOf(['HA', 'dev'])
        .required('Availability mode is required'),
    }).required(),
    onSubmit: async (val) => {
      const submit = async () => {
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
                    nvidiaGpuEnabled: true,
                    instanceType: 'c6a.xlarge',
                  },
                },
                credentialsRef: {
                  name: val.credentialsRef,
                },
                availabilityMode: validateAvailabilityMode(
                  val.availabilityMode
                ),
              },
              metadata: {
                name: val.name,
              },
            },
          });
          if (e) {
            throw e[0];
          }
          toast.success('Cluster created successfully');
          navigate(`/${accountName}/infra/clusters`);
        } catch (err) {
          handleError(err);
        }
      };

      switch (currentStep) {
        case 1:
          nextStep();
          break;
        case 2:
        case 4:
          await submit();
          break;
        default:
          break;
      }
    },
  });

  const getView = () => {
    return (
      <div className="flex flex-col gap-3xl py-3xl">
        <TitleBox
          subtitle="A cluster is a group of interconnected elements working together as a
          single unit."
        />
        <div className="flex flex-col">
          <div className="flex flex-col gap-3xl pb-xl">
            <NameIdView
              nameErrorLabel="isNameError"
              resType="cluster"
              displayName={values.displayName}
              name={values.name}
              label="Cluster name"
              placeholder="Enter cluster name"
              errors={errors.name}
              handleChange={handleChange}
            />
          </div>
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
                mapper(awsRegions, (v) => {
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
              error={!!errors.availabilityMode}
              message={
                errors.availabilityMode ? 'Availability mode is required' : null
              }
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
          <div className="flex flex-row gap-xl justify-start">
            <Button
              variant="primary"
              content="Create"
              suffix={<ArrowRight />}
              type="submit"
              loading={isLoading}
            />
          </div>
        ) : (
          <div className="flex flex-row justify-start">
            <Button
              variant="primary"
              content="Continue"
              suffix={<ArrowRight />}
              type="submit"
            />
          </div>
        )}
      </div>
    );
  };
  return (
    <form
      onSubmit={(e) => {
        if (!values.isNameError) {
          handleSubmit(e);
        } else {
          e.preventDefault();
        }
      }}
    >
      <MultiStepProgressWrapper
        title={
          isOnboarding ? 'Setup your account!' : 'Let’s create new cluster.'
        }
        subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite teams"
        {...(isOnboarding
          ? {}
          : {
              backButton: {
                content: 'Back to clusters',
                to: `/${accountName}/infra/clusters`,
              },
            })}
      >
        <MultiStepProgress.Root
          noJump={isOnboarding}
          currentStep={currentStep}
          jumpStep={jumpStep}
        >
          {!isOnboarding ? (
            <>
              <MultiStepProgress.Step label="Configure cluster" step={1}>
                {getView()}
              </MultiStepProgress.Step>
              <MultiStepProgress.Step label="Review" step={2}>
                <ReviewComponent
                  title="Cluster detail"
                  onEdit={() => {
                    jumpStep(1);
                  }}
                >
                  <div className="flex flex-col p-xl  gap-lg rounded border border-border-default flex-1 overflow-hidden">
                    <div className="flex flex-col gap-md  pb-lg  border-b border-border-default">
                      <div className="bodyMd-semibold text-text-default">
                        Cluster name
                      </div>
                      <div className="bodySm text-text-soft">{values.name}</div>
                    </div>
                    <div className="flex flex-col gap-md  pb-lg  border-b border-border-default">
                      <div className="bodyMd-semibold text-text-default">
                        Cloud provider
                      </div>
                      <div className="bodySm text-text-soft">
                        {values.cloudProvider}
                      </div>
                    </div>
                    {values.cloudProvider === 'aws' && (
                      <div className="flex flex-col gap-md">
                        <div className="bodyMd-semibold text-text-default">
                          Region
                        </div>
                        <div className="bodySm text-text-soft">
                          {values.region}
                        </div>
                      </div>
                    )}
                    <div className="flex flex-col gap-md  pb-lg">
                      <div className="bodyMd-semibold text-text-default">
                        Availability Mode
                      </div>
                      <div className="bodySm text-text-soft">
                        {values.availabilityMode === 'HA'
                          ? 'High Availability'
                          : 'Development'}
                      </div>
                    </div>
                  </div>
                </ReviewComponent>
                <div className="flex flex-row justify-start">
                  <Button
                    loading={isLoading}
                    variant="primary"
                    content="Create"
                    suffix={<ArrowRight />}
                    type="submit"
                  />
                </div>
              </MultiStepProgress.Step>
            </>
          ) : (
            <>
              <MultiStepProgress.Step step={1} label="Create team" />
              <MultiStepProgress.Step
                step={2}
                label="Add your cloud provider"
              />
              <MultiStepProgress.Step
                step={3}
                label="Validate cloud provider"
              />
              <MultiStepProgress.Step step={4} label="Setup first cluster">
                {getView()}
              </MultiStepProgress.Step>
            </>
          )}
        </MultiStepProgress.Root>
      </MultiStepProgressWrapper>
    </form>
  );
};
