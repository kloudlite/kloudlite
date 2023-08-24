import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { Badge } from '~/components/atoms/badge';
import { Button } from '~/components/atoms/button';
import { Checkbox } from '~/components/atoms/checkbox';
import { TextInput } from '~/components/atoms/input';
import Radio from '~/components/atoms/radio';
import Slider from '~/components/atoms/slider';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import { cn } from '~/components/utils';
import AlertDialog from '~/console/components/alert-dialog';
import { IdSelector } from '~/console/components/id-selector';
import RawWrapper from '~/console/components/raw-wrapper';

const ContentWrapper = ({ children }) => (
  <div className="flex flex-col gap-6xl">{children}</div>
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
        <Button content="Back" prefix={ArrowLeft} variant="outline" />
        <Button content="Continue" suffix={ArrowRight} variant="primary" />
      </div>
    </ContentWrapper>
  );
};

const Compute = () => {
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
      <div>
        <Slider step={1} />
      </div>
      <div className="flex flex-row gap-xl justify-end">
        <Button content="Back" prefix={ArrowLeft} variant="outline" />
        <Button content="Continue" suffix={ArrowRight} variant="primary" />
      </div>
    </ContentWrapper>
  );
};
const App = () => {
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
                items={[
                  { label: 'Application details', active: true, id: 1 },
                  {
                    label: 'Compute',
                    active: false,
                    id: 2,
                  },
                  {
                    label: 'Environment',
                    active: false,
                    id: 3,
                  },
                  { label: 'Network', active: false, id: 4 },
                  { label: 'Review', active: false, id: 5 },
                ]}
              />
            </div>

            <Button variant="outline" content="Cancel" size="lg" />
          </>
        }
        rightChildren={
          <div className="flex flex-col gap-6xl">
            <Compute />
          </div>
        }
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

export default App;
