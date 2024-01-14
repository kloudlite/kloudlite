import { useEffect } from 'react';
import {
  NumberInput,
  PasswordInput,
  TextInput,
} from '~/components/atoms/input';
import Slider from '~/components/atoms/slider';
import { TitleBox } from '~/console/components/raw-wrapper';
import { useAppState } from '~/console/page-components/app-states';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';

import { InfoLabel } from '~/console/components/commons';
import { FadeIn, parseValue } from '~/console/page-components/util';
import Select from '~/components/atoms/select';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import Wrapper from '~/console/components/wrapper';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { Button } from '~/components/atoms/button';
import { plans } from '../../../../new-app/datas';

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

const SettingCompute = () => {
  const { app, setApp, getContainer, activeContIndex } = useAppState();
  const { setPerformAction, hasChanges, loading } = useUnsavedChanges();

  const { values, errors, handleChange, submit, resetValues } = useForm({
    initialValues: {
      imageUrl: getContainer(0)?.image || '',
      pullSecret: 'TODO',
      cpuMode: app.metadata?.annotations?.[keyconstants.cpuMode] || 'shared',
      memPerCpu: app.metadata?.annotations?.[keyconstants.memPerCpu] || 1,

      cpu: parseValue(
        app.spec.containers[activeContIndex]?.resourceCpu?.max,
        250
      ),

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
      imageUrl: Yup.string().required(),
      pullSecret: Yup.string(),
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
            [keyconstants.selectedPlan]: val.selectedPlan,
          },
        },
        spec: {
          ...s.spec,
          containers: [
            {
              ...(s.spec.containers?.[0] || {}),
              image: val.imageUrl,
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

  // const getActivePlan = useCallback(() => {
  //   return plans[values.cpuMode as IcpuMode].find(
  //     (v) => v.memoryPerCpu === parseValue(values.selectedPlan, 4)
  //   );
  // }, [values.cpuMode, values.selectedPlan]);

  useEffect(() => {
    submit();
  }, [values]);

  useEffect(() => {
    if (!hasChanges) {
      resetValues();
    }
  }, [hasChanges]);

  return (
    <div>
      <Wrapper
        secondaryHeader={{
          title: 'Compute',
          action: hasChanges && (
            <div className="flex flex-row items-center gap-lg">
              <Button
                disabled={loading}
                variant="basic"
                content="Discard changes"
                onClick={() => setPerformAction('discard-changes')}
              />
              <Button
                disabled={loading}
                content={loading ? 'Committing changes.' : 'View changes'}
                loading={loading}
                onClick={() => setPerformAction('view-changes')}
              />
            </div>
          ),
        }}
      >
        <div className="flex flex-col gap-3xl">
          <TextInput
            label={
              <InfoLabel info="some usefull information" label="Image Url" />
            }
            size="lg"
            value={values.imageUrl}
            onChange={handleChange('imageUrl')}
            error={!!errors.imageUrl}
            message={errors.imageUrl}
          />
          {/* <PasswordInput
            label={
              <InfoLabel info="some usefull information" label="Pull Secret" />
            }
            size="lg"
            value={values.pullSecret}
            // error={!!errors.pullSecret}
            // message={errors.pullSecret}
            // onChange={handleChange('pullSecret')}
          /> */}
        </div>
        {/* <div className="flex flex-col border border-border-default bg-surface-basic-default rounded overflow-hidden">
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
        {/* <div className="flex flex-col gap-md p-2xl rounded border border-border-default bg-surface-basic-default">
        <div className="flex flex-row gap-lg items-center">
          <div className="bodyMd-medium text-text-default">Select CPU</div>
          <code className="bodySm text-text-soft flex-1 text-end">
            {((values.cpu || 1) / 1000).toFixed(2)}vCPU &{' '}
            {(
              ((values.cpu || 1) * parseValue(values.selectedPlan, 4)) /
              1000
            ).toFixed(2)}
            GB Memory
          </code>
        </div>
        <Slider
          step={25}
          min={100}
          max={8000}
          value={values.cpu}
          onChange={(value) => {
            handleChange('cpu')(dummyEvent(value));
          }}
        />
      </div> */}
      </Wrapper>
    </div>
  );
};
export default SettingCompute;
