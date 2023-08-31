import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { Badge } from '~/components/atoms/badge';
import { Button } from '~/components/atoms/button';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ProgressTracker } from '~/components/organisms/progress-tracker';
import { cn } from '~/components/utils';
import AlertDialog from '~/console/components/alert-dialog';
import RawWrapper from '~/console/components/raw-wrapper';
import { useState } from 'react';
import AppEnvironment from './app-environment';
import AppNetwork from './app-network';
import AppReview from './app-review';
import AppDetail from './app-detail';
import AppCompute from './app-compute';

const App = () => {
  const tabs = {
    ENVIRONMENT: 'environment',
    APPLICATION_DETAILS: 'application_details',
    COMPUTE: 'compute',
    NETWORK: 'network',
    REVIEW: 'review',
  };
  const [activeTab, setActiveTab] = useState(tabs.APPLICATION_DETAILS);

  const isActive = (t) => t === activeTab;

  const progressItems = [
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
  ];

  const tab = () => {
    switch (activeTab) {
      case tabs.ENVIRONMENT:
        return <AppEnvironment />;
      case tabs.APPLICATION_DETAILS:
        return <AppDetail />;
      case tabs.COMPUTE:
        return <AppCompute />;
      case tabs.NETWORK:
        return <AppNetwork />;
      case tabs.REVIEW:
        return <AppReview />;
      default:
        return <span>404 | page not found</span>;
    }
  };

  const back = () => {
    const activeTab = progressItems.findIndex((pi) => pi.active);
    if (activeTab !== -1) {
      if (activeTab === 0) {
        // start
      } else {
        setActiveTab(progressItems[activeTab - 1].id);
      }
    }
  };

  const next = () => {
    const activeTab = progressItems.findIndex((pi) => pi.active);
    if (activeTab !== -1) {
      if (activeTab === progressItems.length - 1) {
        // finished
      } else {
        setActiveTab(progressItems[activeTab + 1].id);
      }
    }
  };

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
                items={progressItems}
              />
            </div>

            <Button variant="outline" content="Cancel" size="lg" />
          </>
        }
        rightChildren={
          <>
            {tab()}
            <div className="flex flex-row gap-xl justify-end">
              <Button content="Back" prefix={<ArrowLeft />} variant="outline" onClick={back}/>
              <Button
                content="Continue"
                suffix={<ArrowRight />}
                variant="primary"
                onClick={next}
              />
            </div>
          </>
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

export const handle = {
  noMainLayout: true,
};

export default App;
