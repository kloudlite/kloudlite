import { useMemo } from 'react';
import { toast } from 'react-toastify';
import { NumberInput, TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import Popup from '~/components/molecule/popup';
import { IdSelector } from '~/console/components/id-selector';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsPoolType as awsPoolType } from '~/root/src/generated/gql/server';
import { useOutletContext } from '@remix-run/react';
import { INodepools } from '~/console/server/gql/queries/nodepool-queries';
import Chips from '~/components/atoms/chips';
import { awsRegions } from '~/console/dummy/consts';
import { mapper } from '~/components/utils';
import { findNodePlan, nodePlans, provisionTypes } from './nodepool-utils';
import { IClusterContext } from '../_.$account.$cluster';

interface BaseProps {
  setVisible: (v: boolean) => void;
}

interface Props1 {
  isUpdate: true;
  data: ExtractNodeType<INodepools>;
}

interface Props2 {
  isUpdate?: false;
  data?: ExtractNodeType<INodepools>;
}

type handleProps = BaseProps & (Props1 | Props2);

const Root = ({ setVisible, isUpdate, data }: handleProps) => {
  const api = useConsoleApi();
  const reloadPage = useReload();
  const { cluster } = useOutletContext<IClusterContext>();
  const clusterRegion = cluster.spec?.aws?.region;
  const cloudProvider = cluster.spec?.cloudProvider;

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            displayName: data.displayName,
            name: parseName(data),
            maximum: `${data.spec.maxCount}`,
            minimum: `${data.spec.minCount}`,
            poolType: data.spec.aws?.poolType || 'ec2',
            awsAvailabilityZone:
              data.spec.aws?.availabilityZone ||
              awsRegions.find((v) => v.Name === clusterRegion)?.Zones[0] ||
              '',
            instanceType: data.spec.aws?.ec2Pool?.instanceType || 'c6a.large',

            labels: [],
            taints: [],
          }
        : {
            name: '',
            displayName: '',
            minimum: '1',
            maximum: '1',

            awsAvailabilityZone:
              awsRegions.find((v) => v.Name === clusterRegion)?.Zones[0] || '',

            // onDemand specs
            instanceType: 'c6a.large',

            labels: [],
            taints: [],
          },
      validationSchema: Yup.object({
        name: Yup.string().required('id is required'),
        displayName: Yup.string().required('name is required'),
        minimum: Yup.number(),
        maximum: Yup.number(),
        poolType: Yup.string().required().oneOf(['ec2', 'spot']),
      }),
      onSubmit: async (val) => {
        const getNodeConf = () => {
          const getAwsNodeSpecs = () => {
            switch (val.poolType) {
              case 'ec2':
                return {
                  ec2Pool: {
                    instanceType: val.instanceType,
                    nodes: {},
                  },
                };
              case 'spot':
                const plan = findNodePlan(val.instanceType);
                return {
                  spotPool: {
                    cpuNode: {
                      vcpu: {
                        max: `${plan?.spotSpec.cpuMax}`,
                        min: `${plan?.spotSpec.cpuMin}`,
                      },
                      memoryPerVcpu: {
                        max: `${plan?.spotSpec.memMax}`,
                        min: `${plan?.spotSpec.memMin}`,
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
                  availabilityZone: val.awsAvailabilityZone,
                  nvidiaGpuEnabled: false,
                  poolType: (val.poolType === 'ec2'
                    ? 'ec2'
                    : 'spot') as awsPoolType,
                  ...getAwsNodeSpecs(),
                },
              };
            default:
              return {};
          }
        };

        try {
          if (!isUpdate) {
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
                  ...getNodeConf(),
                },
              },
            });
            if (e) {
              throw e[0];
            }
          } else if (isUpdate) {
            const { errors: e } = await api.updateNodePool({
              clusterName: parseName(cluster),
              pool: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },
                spec: {
                  ...data.spec,
                  maxCount: Number.parseInt(val.maximum, 10),
                  minCount: Number.parseInt(val.minimum, 10),
                  targetCount: Number.parseInt(val.minimum, 10),
                  ...getNodeConf(),
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
            `nodepool ${isUpdate ? 'updated' : 'created'} successfully`
          );
          setVisible(false);
        } catch (err) {
          handleError(err);
        }
      },
    });

  return (
    <form onSubmit={handleSubmit}>
      <Popup.Content>
        <div className="flex flex-col gap-2xl">
          {isUpdate && (
            <Chips.Chip
              {...{
                item: { id: parseName(data) },
                label: parseName(data),
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

          {!isUpdate && (
            <IdSelector
              resType="nodepool"
              onChange={(v) => {
                handleChange('name')(dummyEvent(v));
              }}
              name={values.displayName}
            />
          )}

          {cloudProvider === 'aws' && (
            <>
              <Select
                label="Provision Mode"
                // eslint-disable-next-line react-hooks/rules-of-hooks
                value={useMemo(() => {
                  const mode = provisionTypes.find(
                    (v) => v.value === values.poolType
                  );
                  return mode;
                }, [values.poolType])}
                placeholder="---Select---"
                options={async () => provisionTypes}
                onChange={(value) => {
                  handleChange('poolType')(dummyEvent(value.value));
                }}
              />

              <Select
                label="Availability Zone"
                value={{
                  value: values.awsAvailabilityZone,
                  label: values.awsAvailabilityZone,
                }}
                options={async () =>
                  mapper(
                    awsRegions.find((v) => v.Name === clusterRegion)?.Zones ||
                      [],
                    (v) => ({
                      value: v,
                      label: v,
                    })
                  )
                }
                onChange={(v) => {
                  handleChange('awsAvailabilityZone')(dummyEvent(v.value));
                }}
              />

              <Select
                // eslint-disable-next-line react-hooks/rules-of-hooks
                value={useMemo(() => {
                  const plan = findNodePlan(values.instanceType);
                  return plan;
                }, [values.instanceType])}
                label="Node plan"
                placeholder="---Select---"
                options={async () => nodePlans}
                onChange={(value) => {
                  handleChange('instanceType')(dummyEvent(value.value));
                }}
              />
            </>
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

          {/* {show?.type === DIALOG_TYPE.ADD && ( */}
          {/*   <> */}
          {/*     <Labels */}
          {/*       value={values.labels} */}
          {/*       onChange={(value: any) => */}
          {/*         handleChange('labels')({ target: { value } }) */}
          {/*       } */}
          {/*     /> */}
          {/*     <Taints */}
          {/*       value={[]} */}
          {/*       onChange={(value: any) => */}
          {/*         handleChange('taints')({ target: { value } }) */}
          {/*       } */}
          {/*     /> */}
          {/*   </> */}
          {/* )} */}
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          loading={isLoading}
          type="submit"
          content={isUpdate ? 'Update' : 'Create'}
          variant="primary"
        />
      </Popup.Footer>
    </form>
  );
};

const HandleNodePool = (props: handleProps & { visible: boolean }) => {
  const { isUpdate, data, setVisible, visible } = props;
  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>{isUpdate ? 'Add nodepool' : 'Edit nodepool'}</Popup.Header>

      {(!isUpdate || (isUpdate && data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleNodePool;
