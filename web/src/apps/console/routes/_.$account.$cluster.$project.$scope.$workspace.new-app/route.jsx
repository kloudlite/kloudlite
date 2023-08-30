import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import Slider from '~/components/atoms/slider';
import { Badge } from '~/components/atoms/badge';
import { Button } from '~/components/atoms/button';
import { Checkbox } from '~/components/atoms/checkbox';
import { TextInput } from '~/components/atoms/input';
import Radio from '~/components/atoms/radio';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import { cn } from '~/components/utils';
import AlertDialog from '~/console/components/alert-dialog';
import { IdSelector } from '~/console/components/id-selector';
import RawWrapper from '~/console/components/raw-wrapper';
import { useState } from 'react';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { Chip, ChipGroup, ChipType } from '~/components/atoms/chips';
import HandleConfig from './app-dialogs';

const ContentWrapper = ({ children }) => (
  <div className="flex flex-col gap-6xl w-full">{children}</div>
);

const ApplicationDetail = () => {
  return (
    <ContentWrapper>
      <div className="flex flex-col gap-lg">
        <div className="headingXl text-text-default">Application details</div>
        <div className="bodySm text-text-soft">
          The application streamlines project management through intuitive task
          tracking and collaboration tools.
        </div>
      </div>
      <div className="flex flex-col gap-3xl">
        <TextInput label="Application name" size="lg" />
        <IdSelector name="app" />
        <TextInput label="Description" size="lg" />
      </div>
      <div className="flex flex-row gap-xl justify-end">
        <Button content="Back" prefix={<ArrowLeft />} variant="outline" />
        <Button content="Continue" suffix={<ArrowRight />} variant="primary" />
      </div>
    </ContentWrapper>
  );
};

const Compute = () => {
  const [slidervalue, setSlidervalue] = useState([10]);
  return (
    <ContentWrapper>
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
      <div className="flex flex-row gap-xl justify-end">
        <Button content="Back" prefix={<ArrowLeft />} variant="outline" />
        <Button content="Continue" suffix={<ArrowRight />} variant="primary" />
      </div>
    </ContentWrapper>
  );
};

const Environment = () => {
  const [active, setActive] = useState('environment-variables');
  const [value, setValue] = useState('');
  const [configDialog, setConfigDialog] = useState(null);
  return (
    <ContentWrapper>
      <div className="flex flex-col gap-xl ">
        <div className="headingXl text-text-default">Environment</div>
        <ExtendedFilledTab
          value={active}
          onChange={setActive}
          items={[
            { label: 'Environment variables', to: 'environment-variables' },
            { label: 'Config mount', to: 'config-mount' },
          ]}
        />
      </div>
      <div className="flex flex-col gap-3xl p-3xl rounded border border-border-default">
        <div className="flex flex-row gap-3xl items-center">
          <div className="flex-1">
            <TextInput label="Key" size="lg" />
          </div>
          <div className="flex-1">
            <TextInput
              value={value}
              onChange={({ target }) => setValue(target.value)}
              label="Value"
              size="lg"
              suffix={
                !value ? (
                  <ChipGroup
                    onClick={(e) => {
                      setConfigDialog(true);
                    }}
                  >
                    <Chip
                      label="Config"
                      item={{ name: 'config' }}
                      value="config"
                      type={ChipType.CLICKABLE}
                    />
                    <Chip label="Secrets" type={ChipType.CLICKABLE} />
                  </ChipGroup>
                ) : null
              }
              showclear={value}
            />
          </div>
        </div>
        <div className="flex flex-row gap-md items-center">
          <div className="bodySm text-text-soft flex-1">
            All environment entries be mounted on the path specified in the
            container
          </div>
          <Button content="Add environment" variant="basic" />
        </div>
      </div>
      <div className="flex flex-row gap-xl justify-end">
        <Button content="Back" prefix={<ArrowLeft />} variant="outline" />
        <Button content="Continue" suffix={<ArrowRight />} variant="primary" />
      </div>
      <HandleConfig show={configDialog} setShow={setConfigDialog} />
    </ContentWrapper>
  );
};

const App = () => {
  const tabs = {
    ENVIRONMENT: 'environment',
    APPLICATION_DETAILS: 'application_details',
    COMPUTE: 'compute',
    NETWORK: 'network',
    REVIEW: 'review',
  };
  const [activeTab, setActiveTab] = useState(tabs.APPLICATION_DETAILS);

  const tab = () => {
    switch (activeTab) {
      case tabs.ENVIRONMENT:
        return <Environment />;
      case tabs.APPLICATION_DETAILS:
        return <ApplicationDetail />;
      case tabs.COMPUTE:
        return <Compute />;
      default:
        return <span>404 | page not found</span>;
    }
  };

  const isActive = (t) => t === activeTab;

  return (
    <>
      <RawWrapper
        leftChildren={
          <>
            <BrandLogo detailed={false} size={48} />
            <div className={cn('flex flex-col gap-8xl')}>
              <div className="flex flex-col gap-3xl">
                <div className="text-text-default heading4xl">
                  Let’s create new application.
                </div>
                <div className="text-text-default bodyMd">
                  Create your application under project effortlessly
                </div>
                <div className="flex flex-row gap-md items-center">
                  <Badge>
                    <span className="text-text-strong">Team:</span>
                    <span className="bodySm-semibold text-text-default">
                      xyz
                    </span>
                  </Badge>
                </div>
              </div>
              <ProgressTracker
                onClick={(id) => setActiveTab(id)}
                items={[
                  {
                    label: 'Application details',
                    active: isActive(tabs.APPLICATION_DETAILS),
                    id: tabs.APPLICATION_DETAILS,
                  },
                  {
                    label: 'Compute',
                    active: isActive(tabs.COMPUTE),
                    id: tabs.COMPUTE,
                  },
                  {
                    label: 'Environment',
                    active: isActive(tabs.ENVIRONMENT),
                    id: tabs.ENVIRONMENT,
                  },
                  {
                    label: 'Network',
                    active: isActive(tabs.NETWORK),
                    id: tabs.NETWORK,
                  },
                  {
                    label: 'Review',
                    id: tabs.REVIEW,
                    active: isActive(tabs.REVIEW),
                  },
                ]}
              />
            </div>

            <Button variant="outline" content="Cancel" size="lg" />
          </>
        }
        rightChildren={tab()}
      />

      <AlertDialog
        title="Leave page with unsaved changes?"
        message="Leaving this page will delete all unsaved changes."
        okText="Leave page"
        type="critical"
      />
    </>
  );
};

export const handle = {
  noMainLayout: true,
};

export default App;
