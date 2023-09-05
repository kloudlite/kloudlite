import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import RawWrapper from '~/console/components/raw-wrapper';
import { useState } from 'react';
import { useMapper } from '~/components/utils';
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

  const isActive = (t: string) => t === activeTab;

  const progressItems = [
    {
      label: 'Application details',
      active: isActive(tabs.APPLICATION_DETAILS),
      id: tabs.APPLICATION_DETAILS,
      completed: true,
    },
    {
      label: 'Compute',
      active: isActive(tabs.COMPUTE),
      id: tabs.COMPUTE,
      completed: true,
    },
    {
      label: 'Environment',
      active: isActive(tabs.ENVIRONMENT),
      id: tabs.ENVIRONMENT,
      completed: false,
    },
    {
      label: 'Network',
      active: isActive(tabs.NETWORK),
      id: tabs.NETWORK,
      completed: false,
    },
    {
      label: 'Review',
      id: tabs.REVIEW,
      active: isActive(tabs.REVIEW),
      completed: false,
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

  const items = useMapper(progressItems, (i) => {
    return {
      value: i.id,
      item: {
        ...i,
      },
    };
  });

  return (
    <RawWrapper
      title="Let’s create new application."
      subtitle="Create your application under project effortlessly."
      badgeTitle="Workspace"
      badgeId="WorkspaceId"
      progressItems={items}
      rightChildren={
        <>
          {tab()}
          <div className="flex flex-row gap-xl justify-end">
            <Button
              content="Back"
              prefix={<ArrowLeft />}
              variant="outline"
              onClick={back}
            />
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
  );
};

export const handle = {
  noMainLayout: true,
};

export default App;
