import { useEffect } from 'react';
import { NumberInput } from '~/components/atoms/input';
import Slider from '~/components/atoms/slider';
import { useAppState } from '~/console/page-components/app-states';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { parseValue } from '~/console/page-components/util';
import Select from '~/components/atoms/select';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import Wrapper from '~/console/components/wrapper';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { Button } from '~/components/atoms/button';
import useCustomSwr from '~/lib/client/hooks/use-custom-swr';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName, parseNodes } from '~/console/server/r-utils/common';
import { Checkbox } from '~/components/atoms/checkbox';
import { useMapper } from '~/components/utils';
import { IProjectContext } from '~/console/routes/_main+/$account+/$project+/_layout';
import { useOutletContext } from '@remix-run/react';
import { plans } from '../../../../new-app/datas';
import appInitialFormValues, {
  mapFormValuesToApp,
} from '../../../../new-app/app-utils';

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
  const { app, setApp, getContainer } = useAppState();
  const { setPerformAction, hasChanges, loading } = useUnsavedChanges();
  const { cluster } = useOutletContext<IProjectContext>();

  const api = useConsoleApi();

  const {
    data: nodepoolData,
    isLoading: nodepoolLoading,
    error: nodepoolLoadingError,
  } = useCustomSwr('/nodepools', async () => {
    return api.listNodePools({
      clusterName: parseName(cluster),
      pagination: {
        first: 100,
        orderBy: 'updateTime',
        sortDirection: 'DESC',
      },
    });
  });

  const { values, errors, handleChange, submit, resetValues } = useForm({
    initialValues: appInitialFormValues({
      app,
      getContainer,
    }),
    validationSchema: Yup.object({
      imageUrl: Yup.string().required(),
      pullSecret: Yup.string(),
      cpuMode: Yup.string().required(),
      selectedPlan: Yup.string().required(),
    }),
    onSubmit: (val) => {
      setApp((s) =>
        mapFormValuesToApp({
          appIn: val,
          oldAppIn: s,
        })
      );
    },
  });

  const nodepools = useMapper(parseNodes(nodepoolData), (val) => ({
    label: val.metadata?.name || '',
    value: val.metadata?.name || '',
  }));

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

          <Checkbox
            label="Always pull image on restart"
            checked={values.imagePullPolicy === 'Always'}
            onChange={(val) => {
              const imagePullPolicy = val ? 'Always' : 'IfNotPresent';
              handleChange('imagePullPolicy')(dummyEvent(imagePullPolicy));
            }}
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
              <div className="flex flex-col gap-md p-2xl rounded border border-border-default bg-surface-basic-default">
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
      </Wrapper>
    </div>
  );
};
export default SettingCompute;
