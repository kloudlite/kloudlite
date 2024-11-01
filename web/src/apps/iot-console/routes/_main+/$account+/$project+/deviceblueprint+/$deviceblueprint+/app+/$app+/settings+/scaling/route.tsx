import { useEffect } from 'react';
import Slider from '@kloudlite/design-system/atoms/slider';
import { useAppState } from '~/iotconsole/page-components/app-states';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { Checkbox } from '@kloudlite/design-system/atoms/checkbox';
import AppWrapper from '~/iotconsole/page-components/app/app-wrapper';

const SettingCompute = () => {
  const { app, setApp } = useAppState();

  const { values, handleChange, submit } = useForm({
    initialValues: {
      autoscaling: app.spec.hpa?.enabled || false,
      minReplicas: app.spec.hpa?.minReplicas || 1,
      maxReplicas: app.spec.hpa?.maxReplicas || 3,
      cpuThreshold: app.spec.hpa?.thresholdCpu || 75,
      memoryThreshold: app.spec.hpa?.thresholdMemory || 75,
      replicas: app.spec.replicas || 1,
    },

    validationSchema: Yup.object({
      replicas: Yup.number()
        .transform((value) => (Number.isNaN(value) ? undefined : value))
        .min(1)
        .max(10)
        .test({
          name: 'count',
          message: 'replicas should be from 1 to 10',
          test: (v) => {
            // @ts-ignore
            return !(v > 10 && v < 1);
          },
        })
        .when(['autoscaling'], ([autoscaling], schema) => {
          if (!autoscaling) {
            return schema.required();
          }
          return schema;
        }),
    }),
    onSubmit: (val) => {
      setApp((s) => ({
        ...s,
        metadata: {
          ...s.metadata!,
          annotations: {
            ...(s.metadata?.annotations || {}),
          },
        },
        spec: {
          ...s.spec,
          hpa: {
            enabled: val.autoscaling,
            maxReplicas: val.maxReplicas,
            minReplicas: val.minReplicas,
            thresholdCpu: val.cpuThreshold,
            thresholdMemory: val.memoryThreshold,
          },
          replicas: val.replicas,
        },
      }));
    },
  });

  useEffect(() => {
    submit();
  }, [values]);

  return (
    <AppWrapper title="Scaling">
      <div className="flex flex-col gap-3xl">
        <Checkbox
          label="Autoscaling"
          checked={values.autoscaling}
          onChange={(val) => {
            handleChange('autoscaling')(dummyEvent(val));
          }}
        />
        {values.autoscaling ? (
          <div className="flex flex-col gap-3xl">
            <div className="flex flex-col gap-md p-2xl rounded border border-border-default bg-surface-basic-default">
              <div className="flex flex-row gap-lg items-center">
                <div className="bodyMd-medium text-text-default">
                  Select min and max replicas
                </div>
                <code className="bodyMd text-text-soft flex-1 text-end">
                  {values.minReplicas || 75} min - {values.maxReplicas || 75}{' '}
                  max
                </code>
              </div>
              <Slider
                step={1}
                min={1}
                max={10}
                value={[values.minReplicas, values.maxReplicas]}
                onChange={(value) => {
                  if (Array.isArray(value)) {
                    handleChange('minReplicas')(dummyEvent(value[0]));
                    handleChange('maxReplicas')(dummyEvent(value[1]));
                  }
                }}
              />
            </div>
            <div className="flex flex-row justify-between gap-3xl">
              <div className="flex flex-col gap-md p-2xl rounded border border-border-default bg-surface-basic-default w-full">
                <div className="flex flex-row gap-lg items-center">
                  <div className="bodyMd-medium text-text-default">
                    Select CPU
                  </div>
                  <code className="bodyMd text-text-soft flex-1 text-end">
                    {values.cpuThreshold || 75}% CPU
                  </code>
                </div>
                <Slider
                  step={1}
                  min={50}
                  max={95}
                  value={values.cpuThreshold}
                  onChange={(value) => {
                    handleChange('cpuThreshold')(dummyEvent(value));
                  }}
                />
              </div>
              <div className="flex flex-col gap-md p-2xl rounded border border-border-default bg-surface-basic-default w-full">
                <div className="flex flex-row gap-lg items-center">
                  <div className="bodyMd-medium text-text-default">
                    Select Memory
                  </div>
                  <code className="bodyMd text-text-soft flex-1 text-end">
                    {values.memoryThreshold || 75}% Memory
                  </code>
                </div>
                <Slider
                  step={1}
                  min={50}
                  max={95}
                  value={values.memoryThreshold}
                  onChange={(value) => {
                    handleChange('memoryThreshold')(dummyEvent(value));
                  }}
                />
              </div>
            </div>
          </div>
        ) : (
          <div className="flex flex-col gap-md p-2xl rounded border border-border-default bg-surface-basic-default">
            <div className="flex flex-row gap-lg items-center">
              <div className="bodyMd-medium text-text-default">
                Select replicas
              </div>
              <code className="bodyMd text-text-soft flex-1 text-end">
                {values.replicas || 1} replicas
              </code>
            </div>
            <Slider
              step={1}
              min={1}
              max={10}
              value={values.replicas}
              onChange={(value) => {
                handleChange('replicas')(dummyEvent(value));
              }}
            />
          </div>
        )}
      </div>
    </AppWrapper>
  );
};
export default SettingCompute;
