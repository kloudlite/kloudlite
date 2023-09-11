import RawWrapper from '~/console/components/raw-wrapper';
import { useMapper } from '~/components/utils';
import { useNavigate, useOutletContext } from '@remix-run/react';
import AppEnvironment from './app-environment';
import AppNetwork from './app-network';
import AppReview from './app-review';
import AppDetail from './app-detail';
import AppCompute from './app-compute';
import { AppContextProvider, createAppTabs, useAppState } from './states';
import { FadeIn } from './util';
import { IWorkspaceContext } from '../_.$account.$cluster.$project.$scope.$workspace/route';

const AppComp = () => {
  const { setPage, page, isPageComplete, resetState } = useAppState();
  const isActive = (t: createAppTabs) => t === page;

  const progressItems: {
    label: string;
    id: createAppTabs;
  }[] = [
    {
      label: 'Application details',
      id: 'application_details',
    },
    {
      label: 'Compute',
      id: 'compute',
    },
    {
      label: 'Environment',
      id: 'environment',
    },
    {
      label: 'Network',
      id: 'network',
    },
    {
      label: 'Review',
      id: 'review',
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

  const items = useMapper(progressItems, (i) => {
    return {
      value: i.id,
      item: {
        ...i,
        active: isActive(i.id),
        completed: isPageComplete(i.id),
      },
    };
  });

  const { workspace } = useOutletContext<IWorkspaceContext>();

  const navigate = useNavigate();
  return (
    <RawWrapper
      title="Let’s create new application."
      subtitle="Create your application under project effortlessly."
      badgeTitle={workspace.displayName}
      badgeId={workspace.metadata.name}
      progressItems={items}
      onProgressClick={(p) => {
        if (isPageComplete(p)) setPage(p);
      }}
      onCancel={() => {
        resetState();
        navigate('../apps');
      }}
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
