import { ReactNode, useCallback, useEffect } from 'react';
import { cn } from '~/components/utils';
import FillerAppCompute from '~/iotconsole/assets/app/filler-compute';
import FillerAppDetail from '~/iotconsole/assets/app/filler-details';
import FillerAppEnv from '~/iotconsole/assets/app/filler-env';
import FillerAppNetwork from '~/iotconsole/assets/app/filler-network';
import FillerAppReview from '~/iotconsole/assets/app/filler-review';
import MultiStepProgress, {
  useMultiStepProgress,
} from '~/iotconsole/components/multi-step-progress';
import MultiStepProgressWrapper from '~/iotconsole/components/multi-step-progress-wrapper';
import {
  AppContextProvider,
  useAppState,
} from '~/iotconsole/page-components/app-states';
import AppCompute from '~/iotconsole/page-components/app/compute';
import AppDetail from './app-detail';
import AppEnvironment from './app-environment';
import AppNetwork from './app-network';
import AppReview from './app-review';

const AppComp = () => {
  const { setPage, page, isPageComplete } = useAppState();

  const { currentStep, jumpStep } = useMultiStepProgress({
    defaultStep: page || 1,
    totalSteps: 5,
  });

  useEffect(() => {
    jumpStep(page);
  }, [page]);

  const getFiller = useCallback((): ReactNode => {
    switch (currentStep) {
      case 1:
        return <FillerAppDetail />;
      case 2:
        return <FillerAppCompute />;
      case 3:
        return <FillerAppEnv />;
      case 4:
        return <FillerAppNetwork />;
      case 5:
        return <FillerAppReview />;
      default:
        return null;
    }
  }, [currentStep]);

  return (
    <MultiStepProgressWrapper
      fillerImage={getFiller()}
      title="Let’s create new application."
      subTitle="Create your application under project effortlessly."
      backButton={{
        content: 'Back to apps',
        to: '../apps',
      }}
      className={cn(currentStep === 3 ? 'max-w-[700px]' : '')}
    >
      <MultiStepProgress.Root
        currentStep={currentStep}
        jumpStep={(step) => setPage(step)}
        noJump={(step) => !isPageComplete(step)}
      >
        <MultiStepProgress.Step step={1} label="Application details">
          <AppDetail />
        </MultiStepProgress.Step>
        <MultiStepProgress.Step step={2} label="Compute">
          <AppCompute mode="new" />
        </MultiStepProgress.Step>
        <MultiStepProgress.Step step={3} label="Environment">
          <AppEnvironment />
        </MultiStepProgress.Step>
        <MultiStepProgress.Step step={4} label="Network">
          <AppNetwork />
        </MultiStepProgress.Step>
        <MultiStepProgress.Step step={5} label="Review">
          <AppReview />
        </MultiStepProgress.Step>
      </MultiStepProgress.Root>
    </MultiStepProgressWrapper>
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
