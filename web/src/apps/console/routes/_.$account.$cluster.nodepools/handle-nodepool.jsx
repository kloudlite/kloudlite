import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import SelectInput from '~/components/atoms/select-primitive';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { IdSelector } from '~/console/components/id-selector';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { toast } from 'react-toastify';
import { useOutletContext } from '@remix-run/react';
import { parseName } from '~/console/server/r-urils/common';
import { keyconstants } from '~/console/server/r-urils/key-constants';
import Select from '~/components/atoms/select';
import { useState } from 'react';
import { handleError } from '~/root/lib/utils/common';
import { Labels, Taints } from './taints-and-labels';

const HandleNodePool = ({ show, setShow, cluster }) => {
  const { nodePlans, provisionTypes, spotSpecs } = {
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

  const api = useAPIClient();
  const reloadPage = useReload();

  // @ts-ignore
  const { user } = useOutletContext();
  const { cloudProvider } = cluster.spec;

  const getNodeConf = (val = {}) => {
    const getAwsNodeSpecs = (v) => {
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
            region: cluster.spec.region,
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
        node_type: '',

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
          const nodepool = {
            metadata: {
              name: val.name,
              annotations: {
                [keyconstants.displayName]: val.displayName,
                [keyconstants.author]: user.name,
                [keyconstants.node_type]: val.node_type,
              },
            },
            spec: {
              maxCount: Number.parseInt(val.maximum, 10),
              minCount: Number.parseInt(val.minimum, 10),

              ...getNodeConf(val),
            },
          };

          console.log(nodepool);

          const { errors: e } = await api.createNodePool({
            clusterName: parseName(cluster),
            pool: nodepool,
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

  const [selectedSpotSpec, setSelectedSpotSpec] = useState(null);

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
                <TextInput
                  type="number"
                  label="Capacity"
                  placeholder="Minimum"
                  value={values.minimum}
                  onChange={handleChange('minimum')}
                  prefix="min: "
                />
              </div>
              <div className="flex-1">
                <TextInput
                  type="number"
                  placeholder="Maximum"
                  value={values.maximum}
                  onChange={handleChange('maximum')}
                  prefix="max: "
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
                  <SelectInput.Root
                    value={values.instanceType}
                    label="Node plan"
                    onChange={(e) => {
                      handleChange('instanceType')(e);
                      handleChange('node_type')(e);
                    }}
                  >
                    <SelectInput.Option disabled value="">
                      --Select--
                    </SelectInput.Option>
                    {nodePlans.map((nodeplan) => (
                      <SelectInput.Option
                        key={nodeplan.value}
                        disabled={nodeplan.disabled}
                        value={nodeplan.value}
                      >
                        {nodeplan.label}
                      </SelectInput.Option>
                    ))}
                  </SelectInput.Root>
                )}

                {values.provisionMode === 'spot' && (
                  <Select
                    value={{
                      label: selectedSpotSpec
                        ? selectedSpotSpec.label
                        : undefined,
                      value: selectedSpotSpec
                        ? selectedSpotSpec.value
                        : undefined,
                      spec: selectedSpotSpec,
                    }}
                    label="Spot Specifications"
                    onChange={(value) => {
                      setSelectedSpotSpec(value);
                      handleChange('node_type')(dummyEvent(value));

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
              onChange={(value) =>
                handleChange('labels')({ target: { value } })
              }
            />
            <Taints
              onChange={(value) =>
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
