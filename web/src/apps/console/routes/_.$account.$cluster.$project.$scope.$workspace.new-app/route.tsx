import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import RawWrapper from '~/console/components/raw-wrapper';
import { useMapper } from '~/components/utils';
import { parse } from 'cookie';
import AppEnvironment from './app-environment';
import AppNetwork from './app-network';
import AppReview from './app-review';
import AppDetail from './app-detail';
import AppCompute from './app-compute';
import { AppContextProvider, createAppTabs, useAppState } from './states';
import { FadeIn } from './util';

const AppComp = () => {
  const { app, setPage, page } = useAppState();
  const isActive = (t: createAppTabs) => t === page;

  const progressItems: {
    label: string;
    id: createAppTabs;
    completed: boolean;
  }[] = [
    {
      label: 'Application details',
      id: 'application_details',
      completed: true,
    },
    {
      label: 'Compute',
      id: 'compute',
      completed: true,
    },
    {
      label: 'Environment',
      id: 'environment',
      completed: false,
    },
    {
      label: 'Network',
      id: 'network',
      completed: false,
    },
    {
      label: 'Review',
      id: 'review',
      completed: false,
    },
  ];

  const tab = () => {
    switch (page) {
      case 'application_details':
        return <AppDetail />;
      case 'compute':
        return <AppCompute />;
      case 'environment':
        return <AppEnvironment />;
      case 'network':
        return <AppNetwork />;
      case 'review':
        return <AppReview />;
      default:
        return (
          <FadeIn>
            <span>404 | page not found</span>
          </FadeIn>
        );
    }
  };

  const back = () => {
    const aTab = progressItems.findIndex((pi) => isActive(pi.id));
    if (aTab !== -1) {
      if (aTab === 0) {
        // start
      } else {
        setPage(progressItems[aTab - 1].id);
      }
    }
  };

  const next = () => {
    const activeTab = progressItems.findIndex((pi) => isActive(pi.id));
    if (activeTab !== -1) {
      if (activeTab === progressItems.length - 1) {
        // finished
      } else {
        setPage(progressItems[activeTab + 1].id);
      }
    }
  };

  const items = useMapper(progressItems, (i) => {
    return {
      value: i.id,
      item: {
        ...i,
        active: isActive(i.id),
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
      onProgressClick={setPage}
      // onCancel={page === 'application_details' ? () => {} : undefined}
      rightChildren={tab()}
    />
  );
};

export const handle = {
  noMainLayout: true,
};

const App = () => {
  return (
    <AppContextProvider>
      <AppComp />
    </AppContextProvider>
  );
};

export default App;
