import {
  AppContextProvider,
  useAppState,
} from '~/console/page-components/app-states';
import MultiStepProgress, {
  useMultiStepProgress,
} from '~/console/components/multi-step-progress';
import MultiStepProgressWrapper from '~/console/components/multi-step-progress-wrapper';
import { useEffect } from 'react';
import AppCompute from './app-compute';
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

  return (
    <MultiStepProgressWrapper
      title="Let’s create new application."
      subTitle="Create your application under project effortlessly."
      backButton={{
        content: 'Back to apps',
        to: '../apps',
      }}
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
          <AppCompute />
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
