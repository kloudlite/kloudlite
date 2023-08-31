import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { Checkbox } from '~/components/atoms/checkbox';
import { TextInput } from '~/components/atoms/input';
import Radio from '~/components/atoms/radio';
import Slider from '~/components/atoms/slider';

const AppCompute = () => {
  const [slidervalue, setSlidervalue] = useState([10]);
  return (
    <>
      <div className="flex flex-col gap-lg">
        <div className="headingXl text-text-default">Compute</div>
        <div className="bodySm text-text-soft">
          Compute refers to the processing power and resources used for data
          manipulation and calculations in a system.
        </div>
      </div>
      <div className="flex flex-col gap-3xl">
        <TextInput label="Image Url" size="lg" />
        <TextInput label="Pull secrets" size="lg" />
      </div>
      <div className="flex flex-col border border-border-default rounded overflow-hidden">
        <div className="p-2xl gap-2xl flex flex-row border-b border-border-disabled bg-surface-basic-subdued">
          <div className="flex-1 bodyMd-medium text-text-default">
            Select plan
          </div>
          <Checkbox label="Shared" />
        </div>
        <div className="flex flex-row">
          <div className="flex-1 flex flex-col border-r border-border-disabled">
            <Radio.Root
              withBounceEffect={false}
              className="gap-y-0"
              value="essential-plan"
            >
              <Radio.Item className="p-2xl" value="essential-plan">
                <div className="flex flex-col pl-xl">
                  <div className="headingMd text-text-default">
                    Essential plan
                  </div>
                  <div className="bodySm text-text-soft">
                    The foundational package for your needs.
                  </div>
                </div>
              </Radio.Item>
              <Radio.Item className="p-2xl" value="standard-offerings">
                <div className="flex flex-col pl-xl">
                  <div className="headingMd text-text-default">
                    Standard offerings
                  </div>
                  <div className="bodySm text-text-soft">
                    A well-rounded choice with ample memory.
                  </div>
                </div>
              </Radio.Item>
              <Radio.Item className="p-2xl" value="memory-Boost-package">
                <div className="flex flex-col pl-xl">
                  <div className="headingMd text-text-default">
                    Memory-Boost package
                  </div>
                  <div className="bodySm text-text-soft">
                    High-memory solution for resource-demanding tasks.
                  </div>
                </div>
              </Radio.Item>
            </Radio.Root>
          </div>
          <div className="flex-1 py-2xl">
            <div className="flex flex-row items-center gap-lg py-lg px-2xl">
              <div className="bodyMd-medium text-text-strong flex-1">
                CPU Optimised
              </div>
              <div className="bodyMd text-text-soft">1x (small)</div>
            </div>
            <div className="flex flex-row items-center gap-lg py-lg px-2xl">
              <div className="bodyMd-medium text-text-strong flex-1">
                Compute
              </div>
              <div className="bodyMd text-text-soft">2vCPU</div>
            </div>
            <div className="flex flex-row items-center gap-lg py-lg px-2xl">
              <div className="bodyMd-medium text-text-strong flex-1">
                Memory
              </div>
              <div className="bodyMd text-text-soft">3.75GB</div>
            </div>
          </div>
        </div>
      </div>
      <div className="flex flex-col gap-md p-2xl rounded border border-border-default">
        <div className="flex flex-row gap-lg items-center">
          <div className="bodyMd-medium text-text-default">Select size</div>
          <div className="bodySm text-text-soft flex-1 text-end">
            0.35vCPU & 0.35GB Memory
          </div>
        </div>
        <Slider value={slidervalue} onChange={setSlidervalue} />
      </div>
    </>
  );
};

export default AppCompute;
