import { PasswordInput, TextInput } from '~/components/atoms/input';
import Radio from '~/components/atoms/radio';
import Slider from '~/components/atoms/slider';
import { Button } from '~/components/atoms/button';
import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { useCallback } from 'react';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import { FadeIn, InfoLabel, parseValue } from './util';
import { IcpuMode, plans } from './datas';
import { useAppState } from './states';

const AppCompute = () => {
  const containerIndex = 0;
  const { app, setApp, setPage } = useAppState();
  const { values, errors, handleChange, isLoading, handleSubmit, submit } =
    useForm({
      initialValues: {
        imageUrl: app.spec.containers[containerIndex]?.image || '',
        pullSecret: 'TODO',
        cpuMode: app.metadata.annotations?.[keyconstants.cpuMode] || 'shared',
        selectedPlan:
          app.metadata.annotations?.[keyconstants.selectedPlan] || '4',
        cpu: parseValue(
          app.spec.containers[containerIndex]?.resourceCpu?.max,
          250
        ),
      },
      validationSchema: Yup.object({
        imageUrl: Yup.string().required(),
        pullSecret: Yup.string(),
        cpuMode: Yup.string().required(),
        selectedPlan: Yup.string().required(),
        cpu: Yup.number().required().min(100).max(8000),
      }),
      onSubmit: (val) => {
        setApp((s) => ({
          ...s,
          metadata: {
            ...s.metadata,
            annotations: {
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
                resourceCpu: {
                  max: `${val.cpu}m`,
                  min: `${val.cpu}m`,
                },
                resourceMemory: {
                  max: `${(
                    (values.cpu || 1) * parseValue(values.selectedPlan, 4)
                  ).toFixed(2)}Mi`,
                  min: `${val.cpu}Mi`,
                },
              },
            ],
          },
        }));
      },
    });

  const getActivePlan = useCallback(() => {
    return plans[values.cpuMode as IcpuMode].find(
      (v) => v.memoryPerCpu === parseValue(values.selectedPlan, 4)
    );
  }, [values.cpuMode, values.selectedPlan]);

  return (
    <FadeIn
      onSubmit={(e) => {
        e.preventDefault();

        (async () => {
          const res = await submit();
          if (res) {
            setPage('environment');
          }
        })();
      }}
    >
      <div className="flex flex-col gap-lg">
        <div className="headingXl text-text-default">Compute</div>
        <div className="bodyMd text-text-soft">
          Compute refers to the processing power and resources used for data
          manipulation and calculations in a system.
        </div>
      </div>
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
        <PasswordInput
          label={
            <InfoLabel info="some usefull information" label="Pull Secret" />
          }
          size="lg"
          value={values.pullSecret}
          // error={!!errors.pullSecret}
          // message={errors.pullSecret}
          // onChange={handleChange('pullSecret')}
        />
      </div>
      <div className="flex flex-col border border-border-default rounded overflow-hidden">
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
      </div>
      <div className="flex flex-col gap-md p-2xl rounded border border-border-default">
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
      </div>
      <div className="flex flex-row gap-xl justify-end items-center">
        <Button
          content="App Info"
          prefix={<ArrowLeft />}
          variant="outline"
          onClick={() => {
            (async () => {
              const res = await submit();
              if (res) {
                setPage('application_details');
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

export default AppCompute;
