import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import { NumberInput, TextInput } from '~/components/atoms/input';
import Slider from '~/components/atoms/slider';
import { useAppState } from '~/console/page-components/app-states';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { InfoLabel } from '~/console/components/commons';
import { FadeIn, parseValue } from '~/console/page-components/util';
import Select from '~/components/atoms/select';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { parseNodes } from '~/console/server/r-utils/common';
import useCustomSwr from '~/lib/client/hooks/use-custom-swr';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useMapper } from '~/components/utils';
import {useEffect, useState } from 'react';
import { plans } from './datas';
import {registryHost} from "~/lib/configs/base-url.cjs";

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
  const { app, setApp, setPage, markPageAsCompleted, activeContIndex , getRepoName, getImageTag} =
    useAppState();
  const api = useConsoleApi();

  const {
    data,
    isLoading: repoLoading,
    error: repoLoadingError,
  } = useCustomSwr('/repos', async () => {
    return api.listRepo({});
  });

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
      
      repoName: app.spec.containers[activeContIndex]?.image ? getRepoName(app.spec.containers[activeContIndex]?.image) : '',
      repoImageTag: app.spec.containers[activeContIndex]?.image ? getImageTag(app.spec.containers[activeContIndex]?.image) : '',
      repoAccountName: app.metadata?.annotations?.[keyconstants.repoAccountName] || '',

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
    },
    validationSchema: Yup.object({
      // imageUrl: Yup.string().required(),
      pullSecret: Yup.string(),
      cpuMode: Yup.string().required(),
      selectedPlan: Yup.string().required(),
      // cpu: Yup.number().required().min(100).max(8000),
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
          containers: [
            {
              ...(s.spec.containers?.[0] || {}),
              // image: val.image === '' ? val.repoImageUrl : val.imageUrl,
              image: values.repoAccountName == undefined || values.repoAccountName == '' ? `${values.repoName}:${values.repoImageTag}` : `${registryHost}/${values.repoAccountName}/${values.repoName}:${values.repoImageTag}`,
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
            label="Repo Name"
            size="lg"
            placeholder="Select Repo"
            // value={{ label: '', value: values.repoName }}
              value={
                values.repoName
                ? { label: values.repoName, value: values.repoName }
                :undefined
              }
            // searchable
            creatable={true}
            onChange={(val) => {
              handleChange('repoName')(dummyEvent(val.value));
              if (val.accName == undefined || val.accName == ''){
                handleChange('repoAccountName')(dummyEvent(''));
              }
              else {
                handleChange('repoAccountName')(dummyEvent(val.accName));
              }
            }}
            options={async () => [...repos]}
            error={!!errors.repos || !!repoLoadingError}
            message={
              repoLoadingError ? 'Error fetching repositories.' : errors.app
            }
            loading={repoLoading}
          />

          <Select
            label="Image Tag"
            size="lg"
            placeholder="Select Image Tag"
            // value={{ label: '', value: values.repoImageTag }}
            value={
              values.repoImageTag
                  ? { label: values.repoImageTag, value: values.repoImageTag }
                  :undefined
            }
            creatable={true}
            onChange={(val) => {
              handleChange('repoImageTag')(dummyEvent(val.value));
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
              value={{ label: '', value: values.selectedPlan }}
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

      {/* <div className="flex flex-col border border-border-default rounded overflow-hidden">
        <div className="p-2xl gap-2xl flex flex-row items-center border-b border-border-disabled bg-surface-basic-subdued">
          <div className="flex-1 bodyMd-medium text-text-default">
            Select plan
          </div>
          <ExtendedFilledTab
            size="sm"
            items={[
              {
                value: 'shared',
                label: (
                  <InfoLabel label="Shared" info="some usefull information" />
                ),
              },

              {
                value: 'dedicated',
                label: (
                  <InfoLabel
                    label="Dedicated"
                    info="some usefull information"
                  />
                ),
              },
            ]}
            value={values.cpuMode}
            onChange={(v) => {
              handleChange('cpuMode')(dummyEvent(v));
            }}
          />
        </div>

        <div className="flex flex-row">
          <div className="flex-1 flex flex-col border-r border-border-disabled">
            <Radio.Root
              withBounceEffect={false}
              className="gap-y-0"
              value={values.selectedPlan}
              onChange={(v) => {
                handleChange('selectedPlan')(dummyEvent(v));
              }}
            >
              {[...(plans[values.cpuMode as IcpuMode] || [])].map(
                ({ name, memoryPerCpu, description }) => {
                  return (
                    <Radio.Item
                      key={`${memoryPerCpu}`}
                      className="p-2xl"
                      value={`${memoryPerCpu}`}
                    >
                      <div className="flex flex-col pl-xl">
                        <div className="headingMd text-text-default">
                          {name}
                        </div>
                        <div className="bodySm text-text-soft">
                          {description}
                        </div>
                      </div>
                    </Radio.Item>
                  );
                }
              )}
            </Radio.Root>
          </div>
          {getActivePlan() ? (
            <div className="flex-1 py-2xl">
              <div className="flex flex-row items-center gap-lg py-lg px-2xl">
                <div className="bodyMd-medium text-text-strong flex-1">
                  {getActivePlan()?.name}
                </div>
                <div className="bodyMd text-text-soft">{values.cpuMode}</div>
              </div>
              <div className="flex flex-row items-center gap-lg py-lg px-2xl">
                <div className="bodyMd-medium text-text-strong flex-1">
                  Compute
                </div>
                <div className="bodyMd text-text-soft">{1}vCPU</div>
              </div>
              <div className="flex flex-row items-center gap-lg py-lg px-2xl">
                <div className="bodyMd-medium text-text-strong flex-1">
                  Memory
                </div>
                <div className="bodyMd text-text-soft">
                  {getActivePlan()?.memoryPerCpu}GB
                </div>
              </div>
            </div>
          ) : (
            <div className="flex-1 py-2xl">
              <div className="flex flex-row items-center gap-lg py-lg px-2xl">
                <div className="bodyMd-medium text-text-strong flex-1">
                  Please Select any plan
                </div>
              </div>
            </div>
          )}
        </div>
      </div> */}

      <div className="flex flex-row gap-xl items-center">
        <Button
          content="App Info"
          prefix={<ArrowLeft />}
          variant="outline"
          onClick={() => {
            (async () => {
              const res = await submit();
              if (res) {
                setPage(1);
              }
            })();
          }}
        />

        <div className="text-surface-primary-subdued">|</div>

        <Button
          loading={isLoading}
          type="submit"
          content="Save & Continue"
          suffix={<ArrowRight />}
          variant="primary"
        />
      </div>
    </FadeIn>
  );
};

// const ContainerRepoLayout = () => {
//   const { promise } = useLoaderData<typeof Reposloader>();
//   return (
//       <LoadingComp data={promise}>
//         {({ repository }) => {
//           const repoList = parseNodes(repository);
//           return <AppCompute services={repoList} />;
//         }}
//       </LoadingComp>
//   );
// };
//
// const NewContainerRepo = () => {
//   return <ContainerRepoLayout />;
// };

export default AppCompute;
//  export default NewContainerRepo
