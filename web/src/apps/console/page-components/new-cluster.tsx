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
// import { IAccountContext } from '../routes/_main+/$account+/_layout';
import ProgressWrapper from '../components/progress-wrapper';
import { NameIdView } from '../components/name-id-view';
import { ReviewComponent } from '../routes/_main+/$account+/$project+/$environment+/new-app/app-review';

type props =
  | {
      providerSecrets: IProviderSecrets;
      cloudProvider?: IProviderSecret;
    }
  | {
      providerSecrets?: IProviderSecrets;
      cloudProvider: IProviderSecret;
    };

type steps = 'Configure cluster' | 'Review';

export const NewCluster = ({ providerSecrets, cloudProvider }: props) => {
  const { cloudprovider: cp } = useParams();
  const isOnboarding = !!cp;

  // const [showUnsavedChanges, setShowUnsavedChanges] = useState(false);
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

  // const { account } = useOutletContext<IAccountContext>();
  const [activeState, setActiveState] = useState<steps>('Configure cluster');
  const isActive = (step: steps) => step === activeState;
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

      switch (activeState) {
        case 'Configure cluster':
          setActiveState('Review');
          break;
        case 'Review':
          await submit();
          break;
        default:
          break;
      }
    },
  });

  const getView = () => {
    return (
      <form
        className="flex flex-col gap-3xl py-3xl"
        onSubmit={(e) => {
          if (!values.isNameError) {
            handleSubmit(e);
          } else {
            e.preventDefault();
          }
        }}
      >
        <div className="bodyMd text-text-soft">
          A cluster is a group of interconnected elements working together as a
          single unit.
        </div>
        <div className="flex flex-col">
          <div className="flex flex-col gap-3xl pb-xl">
            <NameIdView
              resType="cluster"
              displayName={values.displayName}
              name={values.name}
              label="Cluster name"
              placeholder="Enter cluster name"
              errors={errors.name}
              onChange={({ name, id }) => {
                handleChange('displayName')(dummyEvent(name));
                handleChange('name')(dummyEvent(id));
              }}
              onCheckError={(check) => {
                handleChange('isNameError')(dummyEvent(check));
              }}
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
              content="Continue"
              suffix={<ArrowRight />}
              type="submit"
            />
          </div>
        ) : (
          <div className="flex flex-row justify-start">
            <Button
              loading={isLoading}
              variant="primary"
              content="Continue"
              suffix={<ArrowRight />}
              type="submit"
            />
          </div>
        )}
      </form>
    );
  };

  const getReviewView = () => {
    return (
      <form onSubmit={handleSubmit} className="flex flex-col gap-3xl py-3xl">
        <ReviewComponent
          title="Cluster detail"
          onEdit={() => {
            setActiveState('Configure cluster');
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
              <div className="flex flex-col gap-md  pb-lg  border-b border-border-default">
                <div className="bodyMd-semibold text-text-default">Region</div>
                <div className="bodySm text-text-soft">{values.region}</div>
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
      </form>
    );
  };

  const items = () => {
    return isOnboarding
      ? [
          {
            label: 'Create Team',
            active: false,
            id: 1,
            completed: true,
          },
          {
            label: 'Add your Cloud Provider',
            active: false,
            id: 3,
            completed: true,
          },
          {
            label: 'Validate Cloud Provider',
            active: false,
            id: 4,
            completed: true,
          },
          {
            label: 'Setup First Cluster',
            active: true,
            id: 5,
            completed: false,
            children: getView(),
          },
        ]
      : [
          {
            label: 'Configure cluster',
            active: isActive('Configure cluster'),
            completed: false,
            children: isActive('Configure cluster') ? getView() : null,
          },
          {
            label: 'Review',
            active: isActive('Review'),
            completed: false,
            children: isActive('Review') ? getReviewView() : null,
          },
        ];
  };

  return (
    <ProgressWrapper
      title={isOnboarding ? 'Setup your account!' : 'Let’s create new cluster.'}
      subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite
  teams"
      progressItems={{
        items: items(),
      }}
      onClick={() => {
        if (isActive('Review')) {
          setActiveState('Configure cluster');
        }
      }}
    />
  );
};
