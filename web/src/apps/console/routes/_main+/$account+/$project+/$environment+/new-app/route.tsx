import { useMapper } from '~/components/utils';
import {
  AppContextProvider,
  createAppTabs,
  useAppState,
} from '~/console/page-components/app-states';
import ProgressWrapper from '~/console/components/progress-wrapper';
import AppCompute from './app-compute';
import AppDetail from './app-detail';
import AppEnvironment from './app-environment';
import AppNetwork from './app-network';
import AppReview from './app-review';
import { FadeIn } from '../../../../../../page-components/util';

const AppComp = () => {
  const { setPage, page, isPageComplete } = useAppState();
  const isActive = (t: createAppTabs) => t === page;

  const progressItems: {
    label: createAppTabs;
  }[] = [
    {
      label: 'Application details',
    },
    {
      label: 'Compute',
    },
    {
      label: 'Environment',
    },
    {
      label: 'Network',
    },
    {
      label: 'Review',
    },
  ];

  const tab = () => {
    switch (page) {
      case 'Application details':
        return <AppDetail />;
      case 'Compute':
        return <AppCompute />;
      case 'Environment':
        return <AppEnvironment />;
      case 'Network':
        return <AppNetwork />;
      case 'Review':
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
      label: i.label,
      active: isActive(i.label),
      completed: isPageComplete(i.label),
      children: isActive(i.label) ? tab() : null,
    };
  });

  return (
    <ProgressWrapper
      onClick={(p) => {
        console.log(p);

        if (isPageComplete(p.label as createAppTabs))
          setPage(p.label as createAppTabs);
      }}
      title="Let’s create new application."
      subTitle="Create your application under project effortlessly."
      progressItems={{
        items,
      }}
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
