import { useEffect, useState } from 'react';
import { toast } from 'react-toastify';
import { NumberInput, TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import Popup from '~/components/molecule/popup';
import { IdSelector } from '~/console/components/id-selector';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsPoolType as awsPoolType } from '~/root/src/generated/gql/server';
import { useOutletContext } from '@remix-run/react';
import { INodepools } from '~/console/server/gql/queries/nodepool-queries';
import { DIALOG_TYPE } from '~/console/utils/commons';
import Chips from '~/components/atoms/chips';
import { nodePlans, provisionTypes, spotSpecs } from './nodepool-utils';
import { Labels, Taints } from './taints-and-labels';
import { IClusterContext } from '../_.$account.$cluster';

const HandleNodePool = ({
  show,
  setShow,
}: IDialog<ExtractNodeType<INodepools> | null, null>) => {
  const [validationSchema, setValidationSchema] = useState(
    Yup.object({
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
    })
  );

  const initialValues = {
    name: '',
    displayName: '',
    minimum: '',
    maximum: '',
    provisionMode: '',

    // onDemand specs
    instanceType: '',

    // spot specs
    cpuMin: '0',
    cpuMax: '0',
    memMin: '0',
    memMax: '0',
    nodeType: '',

    labels: [],
    taints: [],
  };

  const api = useConsoleApi();
  const reloadPage = useReload();
  const { cluster } = useOutletContext<IClusterContext>();
  const cloudProvider = cluster.spec?.cloudProvider;
  const region = cluster.spec?.aws?.region;

  const getNodeConf = (val: typeof initialValues) => {
    const getAwsNodeSpecs = (v: typeof initialValues) => {
      switch (v.provisionMode) {
        case 'on_demand':
          return {
            ec2Pool: {
              instanceType: v.instanceType,
              nodes: {},
            },
          };
        case 'spot':
          return {
            spotPool: {
              cpuNode: {
                vcpu: {
                  max: `${v.cpuMax}`,
                  min: `${v.cpuMin}`,
                },
                memoryPerVcpu: {
                  max: `${v.memMax}`,
                  min: `${v.memMin}`,
                },
              },
              nodes: {},
            },
          };
        default:
          return {};
      }
    };
    switch (cloudProvider) {
      case 'aws':
        return {
          aws: {
            availabilityZone: region || 'ap-south-1a',
            nvidiaGpuEnabled: false,
            poolType: (val.provisionMode === 'on_demand'
              ? 'ec2'
              : 'spot') as awsPoolType,
            ...getAwsNodeSpecs(val),
          },
        };
      default:
        return {};
    }
  };

  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    resetValues,
    isLoading,
    setValues,
  } = useForm({
    initialValues,
    validationSchema,
    onSubmit: async (val) => {
      try {
        if (show?.type === DIALOG_TYPE.ADD) {
          const { errors: e } = await api.createNodePool({
            clusterName: parseName(cluster),
            pool: {
              displayName: val.displayName,
              metadata: {
                name: val.name,
              },
              spec: {
                maxCount: Number.parseInt(val.maximum, 10),
                minCount: Number.parseInt(val.minimum, 10),
                targetCount: Number.parseInt(val.minimum, 10),
                cloudProvider: 'aws',
                ...getNodeConf(val),
              },
            },
          });
          if (e) {
            throw e[0];
          }
        } else if (show?.type === DIALOG_TYPE.EDIT && !!show.data) {
          const { errors: e } = await api.updateNodePool({
            clusterName: parseName(cluster),
            pool: {
              displayName: val.displayName,
              metadata: {
                name: show.data.metadata?.name || '',
              },
              spec: {
                ...show.data.spec,
                maxCount: Number.parseInt(val.maximum, 10),
                minCount: Number.parseInt(val.minimum, 10),
                targetCount: Number.parseInt(val.minimum, 10),
                // cloudProvider: 'aws',
                // aws: {
                //   ec2Pool: show.data.spec.aws?.ec2Pool,
                //   spotPool: show.data.spec.aws?.spotPool,
                //   availabilityZone: region || 'ap-south-1a',
                //   nvidiaGpuEnabled: false,
                //   poolType: show.data.spec.aws?.poolType || 'ec2',
                // },
              },
            },
          });
          if (e) {
            throw e[0];
          }
        }
        reloadPage();
        resetValues();
        toast.success(
          `nodepool ${
            show?.type === DIALOG_TYPE.ADD ? 'created' : 'updated'
          } successfully`
        );
        setShow(null);
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    if (show && show.data && show.type === DIALOG_TYPE.EDIT) {
      setValues((v) => ({
        ...v,
        displayName: show.data?.displayName || '',
        maximum: `${show.data?.spec.maxCount}` || '0',
        minimum: `${show.data?.spec.minCount}` || '0',
      }));
      setValidationSchema(
        // @ts-ignore
        Yup.object({
          displayName: Yup.string().trim().required(),
        })
      );
    }
  }, [show]);

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
            {show?.type === DIALOG_TYPE.EDIT && (
              <Chips.Chip
                {...{
                  item: { id: parseName(show.data) },
                  label: parseName(show.data),
                  prefix: 'Id:',
                  disabled: true,
                  type: 'BASIC',
                }}
              />
            )}
            <TextInput
              label="Name"
              value={values.displayName}
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
            />
            {show?.type === DIALOG_TYPE.ADD && (
              <IdSelector
                resType="nodepool"
                onChange={(v) => {
                  handleChange('name')(dummyEvent(v));
                }}
                name={values.displayName}
              />
            )}
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
                    options={async () => provisionTypes}
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
                    options={async () => nodePlans}
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
                    options={async () =>
                      spotSpecs.map((spec) => ({
                        label: spec.label,
                        value: spec.value,
                        spec,
                      }))
                    }
                  />
                )}
              </>
            )}

            {show?.type === DIALOG_TYPE.ADD && (
              <>
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
              </>
            )}
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
