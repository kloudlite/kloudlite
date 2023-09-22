import React, { ReactElement, useState } from 'react';
import { ChildrenProps } from '~/components/types';

interface IUseMultiStep {
  defaultStep: number;
  totalSteps: number;
}
export const useMultiStep = ({
  defaultStep = 1,
  totalSteps = 1,
}: IUseMultiStep) => {
  const [currentIndex, setCurrentIndex] = useState(defaultStep);

  const onNext = () => {
    if (currentIndex < totalSteps) {
      setCurrentIndex((prev) => prev + 1);
    }
  };
  const onPrevious = () => {
    if (currentIndex > 1) {
      setCurrentIndex((prev) => prev - 1);
    }
  };
  return { currentStep: currentIndex, onNext, onPrevious };
};

interface IStep extends ChildrenProps {
  step: number;
}
const Step = ({ children, step }: IStep) => {
  return <div>{children}</div>;
};

interface IRoot extends ChildrenProps {
  currentStep: number;
}
const Root = ({ children, currentStep }: IRoot) => {
  return (
    <div>
      {React.Children.map(children as ReactElement[], (child) => {
        if (child?.props?.step === currentStep) {
          return child;
        }
        return null;
      })}
    </div>
  );
};

const MultiStep = {
  Root,
  Step,
};

export default MultiStep;
