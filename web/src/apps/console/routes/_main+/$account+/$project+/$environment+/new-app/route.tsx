import { Folders } from '@jengaicons/react';
import { useNavigate, useOutletContext } from '@remix-run/react';
import { useMapper } from '~/components/utils';
import RawWrapper from '~/console/components/raw-wrapper';
import {
  AppContextProvider,
  createAppTabs,
  useAppState,
} from '~/console/page-components/app-states';
import { parseName } from '~/console/server/r-utils/common';
import ProgressWrapper from '~/console/components/progress-wrapper';
import { useEffect } from 'react';
import AppCompute from './app-compute';
import AppDetail from './app-detail';
import AppEnvironment from './app-environment';
import AppNetwork from './app-network';
import AppReview from './app-review';
import { FadeIn } from '../../../../../../page-components/util';
import { IEnvironmentContext } from '../_layout';

const AppComp = () => {
  const { setPage, page, isPageComplete, resetState } = useAppState();
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

  const { environment } = useOutletContext<IEnvironmentContext>();

  const navigate = useNavigate();

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
  return (
    <RawWrapper
      title="Let’s create new application."
      subtitle="Create your application under project effortlessly."
      badge={{
        title: environment.displayName,
        subtitle: parseName(environment),
        image: <Folders size={20} />,
      }}
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
