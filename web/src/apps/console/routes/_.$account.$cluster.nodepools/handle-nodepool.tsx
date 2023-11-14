import { useState } from 'react';
import { toast } from 'react-toastify';
import { NumberInput, TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import Popup from '~/components/molecule/popup';
import { IdSelector } from '~/console/components/id-selector';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ICluster } from '~/console/server/gql/queries/cluster-queries';
import {
  parseName,
  validateProvisionMode,
} from '~/console/server/r-utils/common';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { nodePlans, provisionTypes, spotSpecs } from './nodepool-utils';
import { Labels, Taints } from './taints-and-labels';

const HandleNodePool = ({
  show,
  setShow,
  cluster,
}: IDialog<any> & { cluster: ICluster }) => {
  const initialValues = {
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
  };

  const api = useConsoleApi();
  const reloadPage = useReload();

  const cloudProvider = cluster.spec?.cloudProvider;

  const getNodeConf = (val: typeof initialValues) => {
    const getAwsNodeSpecs = (v: typeof initialValues) => {
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
            region: cluster.spec?.aws?.region || '',
            vpc: '',
            provisionMode: validateProvisionMode(val.provisionMode),
            ...getAwsNodeSpecs(val),
          },
        };
      default:
        return {};
    }
  };

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues,
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
              displayName: val.displayName,
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
          setShow(null);
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

  const [selectedProvisionMode, setSelectedProvisionMode] = useState<
    (typeof provisionTypes)[number] | null
  >(null);

  const [selectedNodePlan, setSelectedNodePlan] = useState<{
    label: string;
    value: string;
  } | null>(null);

  return (
    <Popup.Root
      show={show as any}
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
                  <Select
                    label="Provision Mode"
                    value={selectedProvisionMode || undefined}
                    placeholder="---Select---"
                    options={provisionTypes}
                    onChange={(value) => {
                      setSelectedProvisionMode(value);
                      handleChange('provisionMode')(dummyEvent(value.value));
                    }}
                  />
                )}
                {values.provisionMode === 'on_demand' && (
                  <Select
                    value={selectedNodePlan || undefined}
                    label="Node plan"
                    placeholder="---Select---"
                    options={nodePlans}
                    onChange={(value) => {
                      setSelectedNodePlan(value);
                      handleChange('instanceType')({
                        target: { value: value.value },
                      });
                      handleChange('nodeType')({
                        target: { value: value.value },
                      });
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
                    placeholder="---Select---"
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
              value={[]}
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
