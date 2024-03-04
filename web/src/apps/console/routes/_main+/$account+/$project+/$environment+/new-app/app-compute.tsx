import { NumberInput } from '~/components/atoms/input';
import Slider from '~/components/atoms/slider';
import { useAppState } from '~/console/page-components/app-states';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { FadeIn, parseValue } from '~/console/page-components/util';
import Select from '~/components/atoms/select';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import {parseName, parseNodes} from '~/console/server/r-utils/common';
import useCustomSwr from '~/lib/client/hooks/use-custom-swr';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useMapper } from '~/components/utils';
import { registryHost } from '~/lib/configs/base-url.cjs';
import { BottomNavigation } from '~/console/components/commons';
import { plans } from './datas';
import {useOutletContext } from '@remix-run/react';
import {IAppContext} from "~/console/routes/_main+/$account+/$project+/$environment+/app+/$app+/_layout";

const valueRender = ({
  label,
  isShared,
  memoryPerCpu,
}: {
  label: string;
  isShared: boolean;
  memoryPerCpu: number;
}) => {
  return (
    <div className="flex flex-row gap-xl items-center w-full justify-between bodyMd text-text-default">
      <span className="flex flex-row items-center gap-xl">
        <span>{label}</span>-<span>{isShared ? 'Shared' : 'Dedicated'}</span>
      </span>
      <span className="flex flex-row items-center gap-xl">
        <span>{memoryPerCpu}GB/vCPU</span>
      </span>
    </div>
  );
};

const AppCompute = () => {
  const {
    app,
    setApp,
    setPage,
    markPageAsCompleted,
    activeContIndex,
    getRepoName,
    getImageTag,
  } = useAppState();
  const api = useConsoleApi();
  const {cluster} = useOutletContext<IAppContext>()


  const {
    data,
    isLoading: repoLoading,
    error: repoLoadingError,
  } = useCustomSwr('/repos', async () => {
    return api.listRepo({});
  });

  const {
    data: nodepoolData,
    isLoading: nodepoolLoading,
    error: nodepoolLoadingError,
  } = useCustomSwr('/nodepools', async () => {
    return api.listNodePools({clusterName: parseName(cluster)})
  })

  const { values, errors, handleChange, isLoading, submit } = useForm({
    initialValues: {
      imageUrl: app.spec.containers[activeContIndex]?.image || '',
      pullSecret: 'TODO',
      cpuMode: app.metadata?.annotations?.[keyconstants.cpuMode] || 'shared',
      memPerCpu: app.metadata?.annotations?.[keyconstants.memPerCpu] || '1',

      cpu: parseValue(
        app.spec.containers[activeContIndex]?.resourceCpu?.max,
        250
      ),

      repoName: app.spec.containers[activeContIndex]?.image
        ? getRepoName(app.spec.containers[activeContIndex]?.image)
        : '',
      repoImageTag: app.spec.containers[activeContIndex]?.image
        ? getImageTag(app.spec.containers[activeContIndex]?.image)
        : '',
      repoAccountName:
        app.metadata?.annotations?.[keyconstants.repoAccountName] || '',

      selectedPlan:
        app.metadata?.annotations[keyconstants.selectedPlan] || 'shared-1',
      selectionMode:
        app.metadata?.annotations[keyconstants.selectionModeKey] || 'quick',
      manualCpuMin: parseValue(
        app.spec.containers[activeContIndex].resourceCpu?.min,
        0
      ),
      manualCpuMax: parseValue(
        app.spec.containers[activeContIndex].resourceCpu?.max,
        0
      ),
      manualMemMin: parseValue(
        app.spec.containers[activeContIndex].resourceMemory?.min,
        0
      ),
      manualMemMax: parseValue(
        app.spec.containers[activeContIndex].resourceMemory?.max,
        0
      ),

      nodepoolName: app.spec.nodeSelector?.[keyconstants.nodepoolName] || ''
    },
    validationSchema: Yup.object({
      pullSecret: Yup.string(),
      repoName: Yup.string().required(),
      repoImageTag: Yup.string().required(),
      cpuMode: Yup.string().required(),
      selectedPlan: Yup.string().required(),
    }),
    onSubmit: (val) => {
      setApp((s) => ({
        ...s,
        metadata: {
          ...s.metadata!,
          annotations: {
            ...(s.metadata?.annotations || {}),
            [keyconstants.cpuMode]: val.cpuMode,
            [keyconstants.memPerCpu]: val.memPerCpu,
            [keyconstants.selectionModeKey]: val.selectionMode,
            [keyconstants.selectedPlan]: val.selectedPlan,
            [keyconstants.repoAccountName]: val.repoAccountName,
          },
        },
        spec: {
          ...s.spec,
          nodeSelector: {
            ...(s.spec.nodeSelector || {}),
            [keyconstants.nodepoolName]: val.nodepoolName
          },
          containers: [
            {
              ...(s.spec.containers?.[0] || {}),
              // image: val.image === '' ? val.repoImageUrl : val.imageUrl,
              image:
                values.repoAccountName === undefined ||
                values.repoAccountName === ''
                  ? `${values.repoName}:${values.repoImageTag}`
                  : `${registryHost}/${values.repoAccountName}/${values.repoName}:${values.repoImageTag}`,
              name: 'container-0',
              resourceCpu:
                val.selectionMode === 'quick'
                  ? {
                      max: `${val.cpu}m`,
                      min: `${val.cpu}m`,
                    }
                  : {
                      max: `${val.manualCpuMax}m`,
                      min: `${val.manualCpuMin}m`,
                    },
              resourceMemory:
                val.selectionMode === 'quick'
                  ? {
                      max: `${(
                        (values.cpu || 1) * parseValue(values.memPerCpu, 4)
                      ).toFixed(2)}Mi`,
                      min: `${val.cpu}Mi`,
                    }
                  : {
                      max: `${val.manualMemMax}Mi`,
                      min: `${val.manualMemMin}Mi`,
                    },
            },
          ],
        },
      }));
    },
  });

  const repos = useMapper(parseNodes(data), (val) => ({
    label: val.name,
    value: val.name,
    accName: val.accountName,
  }));

  const nodepools = useMapper(parseNodes(nodepoolData), (val) => ({
    label: val.metadata?.name || '',
    value: val.metadata?.name || '',
  }))

  const {
    data: digestData,
    isLoading: digestLoading,
    error: digestError,
  } = useCustomSwr(
    () => `/digests_${values.repoName}`,
    async () => {
      return api.listDigest({ repoName: values.repoName });
    }
  );

  return (
    <FadeIn
      onSubmit={(e) => {
        e.preventDefault();

        (async () => {
          const res = await submit();
          if (res) {
            setPage(3);
            markPageAsCompleted(2);
          }
        })();
      }}
    >
      <div className="bodyMd text-text-soft">
        Compute refers to the processing power and resources used for data
        manipulation and calculations in a system.
      </div>
      <div className="flex flex-col gap-3xl">

        <Select
            label="Nodepool Name"
            size="lg"
            placeholder="Select Nodepool"
            value={values.nodepoolName}
            creatable
            onChange={(val) => {
              handleChange('nodepoolName')(dummyEvent(val.value));
            }}
            options={async () => [...nodepools]}
            error={!!errors.repos || !!nodepoolLoadingError}
            message={
              nodepoolLoadingError ? 'Error fetching nodepools.' : errors.app
            }
            loading={nodepoolLoading}
        />

        <Select
          label="Repo Name"
          size="lg"
          placeholder="Select Repo"
          value={values.repoName}
          creatable
          onChange={(val) => {
            handleChange('repoName')(dummyEvent(val.value));
            if (val.accName === undefined || val.accName === '') {
              handleChange('repoAccountName')(dummyEvent(''));
            } else {
              handleChange('repoAccountName')(dummyEvent(val.accName));
            }
          }}
          options={async () => [...repos]}
          error={!!errors.repoName || !!repoLoadingError}
          message={
            repoLoadingError ? 'Error fetching repositories.' : errors.app
          }
          loading={repoLoading}
        />

        <Select
          label="Image Tag"
          size="lg"
          placeholder="Select Image Tag"
          value={values.repoImageTag}
          creatable
          onChange={(_, val) => {
            handleChange('repoImageTag')(dummyEvent(val));
          }}
          options={async () =>
            [
              ...new Set(
                parseNodes(digestData)
                  .map((item) => item.tags)
                  .flat()
              ),
            ].map((item) => ({
              label: item,
              value: item,
            }))
          }
          error={!!errors.repoImageTag || !!digestError}
          message={
            // eslint-disable-next-line no-nested-ternary
            errors.repoImageTag
              ? errors.repoImageTag
              : digestError
              ? 'Failed to load Image tags.'
              : ''
          }
          loading={digestLoading}
        />
      </div>

      <div className="flex flex-col">
        <div className="flex flex-row gap-lg items-center pb-3xl">
          <div className="flex-1">
            <ExtendedFilledTab
              value={values.selectionMode}
              onChange={(e) => {
                handleChange('selectionMode')(dummyEvent(e));
              }}
              items={[
                { label: 'Quick', value: 'quick' },
                {
                  label: 'Manual',
                  value: 'manual',
                },
              ]}
            />
          </div>
        </div>
        {values.selectionMode === 'quick' ? (
          <div className="flex flex-col gap-3xl">
            <Select
              value={values.selectedPlan}
              label="Plan"
              size="lg"
              options={async () => [
                ...Object.entries(plans).map(([_, vs]) => ({
                  label: vs.label,
                  options: vs.options.map((op) => ({
                    ...op,
                    render: () => (
                      <div className="flex flex-row justify-between">
                        <div>{op.label}</div>
                        <div className="bodySm">{op.memoryPerCpu}GB/vCPU</div>
                      </div>
                    ),
                  })),
                })),
              ]}
              valueRender={valueRender}
              onChange={(v) => {
                handleChange('selectedPlan')(dummyEvent(v.value));
                handleChange('memPerCpu')(dummyEvent(v.memoryPerCpu));
                handleChange('cpuMode')(
                  dummyEvent(v.isShared ? 'shared' : 'dedicated')
                );
              }}
            />
            <div className="flex flex-col gap-md p-2xl rounded border border-border-default">
              <div className="flex flex-row gap-lg items-center">
                <div className="bodyMd-medium text-text-default">
                  Select CPU
                </div>
                <code className="bodyMd text-text-soft flex-1 text-end">
                  {((values.cpu || 1) / 1000).toFixed(2)}vCPU &{' '}
                  {(
                    ((values.cpu || 1) * parseValue(values.memPerCpu, 4)) /
                    1000
                  ).toFixed(2)}
                  GB Memory
                </code>
              </div>
              <Slider
                step={100}
                min={100}
                max={8000}
                value={values.cpu}
                onChange={(value) => {
                  handleChange('cpu')(dummyEvent(value));
                }}
              />
            </div>
          </div>
        ) : (
          <div className="flex flex-col gap-3xl">
            <div className="flex flex-col gap-md">
              <div className="flex flex-row items-start gap-2xl">
                <div className="basis-full">
                  <NumberInput
                    value={values.manualCpuMin}
                    onChange={handleChange('manualCpuMin')}
                    label="CPU request"
                    size="lg"
                    suffix="m"
                    extra={
                      <span className="bodySm text-text-soft">
                        1000m = 1VCPU
                      </span>
                    }
                  />
                </div>
                <div className="basis-full">
                  <NumberInput
                    value={values.manualCpuMax}
                    onChange={handleChange('manualCpuMax')}
                    label="CPU limit"
                    extra={
                      <span className="bodySm text-text-soft">
                        1000m = 1VCPU
                      </span>
                    }
                    size="lg"
                    suffix="m"
                  />
                </div>
              </div>
            </div>
            <div className="flex flex-col gap-md">
              <div className="flex flex-row items-start gap-2xl">
                <div className="basis-full">
                  <NumberInput
                    value={values.manualMemMin}
                    onChange={handleChange('manualMemMin')}
                    label="Memory request"
                    size="lg"
                    suffix="MB"
                  />
                </div>
                <div className="basis-full">
                  <NumberInput
                    value={values.manualMemMax}
                    onChange={handleChange('manualMemMax')}
                    label="Memory limit"
                    size="lg"
                    suffix="MB"
                  />
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      <BottomNavigation
        primaryButton={{
          loading: isLoading,
          type: 'submit',
          content: 'Save & Continue',
          variant: 'primary',
        }}
        secondaryButton={{
          content: 'App Info',
          variant: 'outline',
          onClick: () => {
            (async () => {
              const res = await submit();
              if (res) {
                setPage(1);
              }
            })();
          },
        }}
      />
    </FadeIn>
  );
};
export default AppCompute;
