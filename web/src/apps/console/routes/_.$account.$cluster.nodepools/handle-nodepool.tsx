import { NumberInput, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import SelectInput from '~/components/atoms/select-primitive';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { IdSelector } from '~/console/components/id-selector';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { toast } from 'react-toastify';
import { parseName } from '~/console/server/r-urils/common';
import { keyconstants } from '~/console/server/r-urils/key-constants';
import Select from '~/components/atoms/select';
import { useState } from 'react';
import { handleError } from '~/root/lib/utils/common';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IHandleProps } from '~/console/server/utils/common';
import { ICluster } from '~/console/server/gql/queries/cluster-queries';
import { NodePoolIn } from '~/root/src/generated/gql/server';
import { NN } from '~/root/lib/types/common';
import { Labels, Taints } from './taints-and-labels';

const HandleNodePool = ({
  show,
  setShow,
  cluster,
}: IHandleProps & { cluster: ICluster }) => {
  type IAWSNodeConfig = NN<NodePoolIn['spec']['awsNodeConfig']>;
  type IProvisionMode = NN<IAWSNodeConfig['provisionMode']>;
  type ISpotSpec = NN<IAWSNodeConfig['spotSpecs']> & {
    label: string;
    value: string;
    disabled: boolean;
  };

  const {
    nodePlans,
    provisionTypes,
    spotSpecs,
  }: {
    nodePlans: { label: string; value: string; disabled: boolean }[];
    provisionTypes: { label: string; value: IProvisionMode }[];
    spotSpecs: ISpotSpec[];
  } = {
    spotSpecs: [
      {
        cpuMax: 4,
        cpuMin: 4,
        memMax: 8192,
        memMin: 8192,
        disabled: false,
        label: '1x - small - 2VCPU 3.75GB Memory',
        value: 'id',
      },
    ],
    nodePlans: [
      {
        label: 'CPU Optimised',
        value: 'CPU Optimised',
        disabled: true,
      },
      {
        label: '1x - small - 2VCPU 3.75GB Memory',
        value: 'c6a-large',
        disabled: false,
      },
    ],
    provisionTypes: [
      { label: 'On-Demand', value: 'on_demand' },
      { label: 'Spot 70% discount', value: 'spot' },
    ],
  };

  const api = useConsoleApi();
  const reloadPage = useReload();

  const cloudProvider = cluster.spec?.cloudProvider;

  const getNodeConf = (val: any) => {
    const getAwsNodeSpecs = (v: any) => {
      switch (v.provisionMode) {
        case 'on_demand':
          return {
            onDemandSpecs: {
              instanceType: v.instanceType,
            },
          };
        case 'spot':
          return {
            spotSpecs: {
              cpuMax: v.cpuMax,
              cpuMin: v.cpuMin,
              memMax: v.memMax,
              memMin: v.memMin,
            },
          };
        default:
          return {};
      }
    };
    switch (cloudProvider) {
      case 'aws':
        return {
          awsNodeConfig: {
            region: cluster.spec?.region,
            vpc: '',
            provisionMode: val.provisionMode,
            ...getAwsNodeSpecs(val),
          },
        };
      default:
        return {};
    }
  };

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        name: '',
        displayName: '',
        minimum: '',
        maximum: '',
        provisionMode: '',

        // onDemand specs
        instanceType: '',

        // spot specs
        cpuMin: 0,
        cpuMax: 0,
        memMin: 0,
        memMax: 0,
        nodeType: '',

        labels: [],
        taints: [],
      },
      validationSchema: Yup.object({
        name: Yup.string().required('id is required'),
        displayName: Yup.string().required('name is required'),
        minimum: Yup.number(),
        maximum: Yup.number(),
        provisionMode: Yup.string().required().oneOf(['on_demand', 'spot']),

        // spot specs
        cpuMax: Yup.number(),
        cpuMin: Yup.number(),
        memMax: Yup.number(),
        memMin: Yup.number(),
      }),
      onSubmit: async (val) => {
        try {
          const { errors: e } = await api.createNodePool({
            clusterName: parseName(cluster),
            pool: {
              metadata: {
                name: val.name,
                annotations: {
                  [keyconstants.nodeType]: val.nodeType,
                },
              },
              spec: {
                maxCount: Number.parseInt(val.maximum, 10),
                minCount: Number.parseInt(val.minimum, 10),
                targetCount: Number.parseInt(val.minimum, 10),
                ...getNodeConf(val),
              },
            },
          });
          if (e) {
            throw e[0];
          }
          reloadPage();
          resetValues();
          toast.success('nodepool created successfully');
          setShow(false);
        } catch (err) {
          handleError(err);
        }
      },
    });

  const [selectedSpotSpec, setSelectedSpotSpec] = useState<{
    label: string;
    value: string;
    spec: (typeof spotSpecs)[number];
  } | null>(null);

  return (
    <Popup.Root
      show={show}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }

        setShow(e);
      }}
    >
      <Popup.Header>
        {show?.type === 'add' ? 'Add nodepool' : 'Edit nodepool'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <TextInput
              label="Name"
              value={values.displayName}
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
            />
            <IdSelector
              resType="nodepool"
              onChange={(v) => {
                handleChange('name')(dummyEvent(v));
              }}
              name={values.displayName}
            />
            <div className="flex flex-row gap-xl items-end">
              <div className="flex-1">
                <NumberInput
                  label="Capacity"
                  placeholder="Minimum"
                  value={values.minimum}
                  onChange={handleChange('minimum')}
                />
              </div>
              <div className="flex-1">
                <NumberInput
                  placeholder="Maximum"
                  value={values.maximum}
                  onChange={handleChange('maximum')}
                />
              </div>
            </div>

            {cloudProvider === 'aws' && (
              <>
                {show?.type === 'add' && (
                  <SelectInput.Root
                    value={values.provisionMode}
                    label="Provision Mode"
                    onChange={handleChange('provisionMode')}
                  >
                    <SelectInput.Option disabled value="">
                      --Select--
                    </SelectInput.Option>
                    {provisionTypes.map(({ label, value }) => (
                      <SelectInput.Option value={value} key={value}>
                        {label}
                      </SelectInput.Option>
                    ))}
                  </SelectInput.Root>
                )}

                {values.provisionMode === 'on_demand' && (
                  // <SelectInput.Root
                  //   value={values.instanceType}
                  //   label="Node plan"
                  //   onChange={(e) => {
                  //     handleChange('instanceType')(e);
                  //     handleChange('nodeType')(e);
                  //   }}
                  // >
                  //   <SelectInput.Option disabled value="">
                  //     --Select--
                  //   </SelectInput.Option>
                  //   {nodePlans.map((nodeplan) => (
                  //     <SelectInput.Option
                  //       key={nodeplan.value}
                  //       disabled={nodeplan.disabled}
                  //       value={nodeplan.value}
                  //     >
                  //       {nodeplan.label}
                  //     </SelectInput.Option>
                  //   ))}
                  // </SelectInput.Root>
                  <Select
                    value={undefined}
                    options={nodePlans}
                    onChange={({ label }) => {
                      console.log(label);
                    }}
                  />
                )}

                {values.provisionMode === 'spot' && (
                  <Select
                    value={
                      selectedSpotSpec
                        ? {
                            label: selectedSpotSpec.label,
                            value: selectedSpotSpec.value,
                            spec: selectedSpotSpec.spec,
                          }
                        : undefined
                    }
                    label="Spot Specifications"
                    onChange={(value) => {
                      setSelectedSpotSpec(value);
                      handleChange('nodeType')(dummyEvent(value));

                      handleChange('cpuMax')(dummyEvent(value.spec.cpuMax));
                      handleChange('cpuMin')(dummyEvent(value.spec.cpuMin));
                      handleChange('memMax')(dummyEvent(value.spec.memMax));
                      handleChange('memMin')(dummyEvent(value.spec.memMin));
                    }}
                    options={spotSpecs.map((spec) => ({
                      label: spec.label,
                      value: spec.value,
                      spec,
                    }))}
                  />
                )}
              </>
            )}

            <Labels
              value={values.labels}
              onChange={(value: any) =>
                handleChange('labels')({ target: { value } })
              }
            />
            <Taints
              onChange={(value: any) =>
                handleChange('taints')({ target: { value } })
              }
            />
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button closable content="Cancel" variant="basic" />
          <Popup.Button
            loading={isLoading}
            type="submit"
            content="Save"
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

export default HandleNodePool;
