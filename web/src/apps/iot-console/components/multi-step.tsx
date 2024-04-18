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
  const reset = () => {
    setCurrentIndex(1);
  };
  return { currentStep: currentIndex, onNext, onPrevious, reset };
};

interface IStep extends ChildrenProps {
  step: number;
  className?: string;
}
const Step = ({ children, step, className }: IStep) => {
  return (
    <div className={className} data-step={step}>
      {children}
    </div>
  );
};

interface IRoot extends ChildrenProps {
  currentStep: number;
}
const Root = ({ children, currentStep }: IRoot) => {
  return (
    <>
      {React.Children.map(children as ReactElement[], (child) => {
        if (child?.props?.step === currentStep) {
          return child;
        }
        return null;
      })}
    </>
  );
};

const MultiStep = {
  Root,
  Step,
};

export default MultiStep;
