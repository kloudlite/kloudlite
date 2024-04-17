import { Check } from '~/console/components/icons';
import React, { Children, ReactElement, ReactNode, useState } from 'react';
import { cn } from '~/components/utils';

interface IUseMultiStepProgress {
  defaultStep: number;
  totalSteps: number;
  onChange?: (step: number) => void;
}
export const useMultiStepProgress = ({
  defaultStep = 1,
  totalSteps = 1,
  onChange,
}: IUseMultiStepProgress) => {
  const [currentIndex, setCurrentIndex] = useState(defaultStep);

  const nextStep = () => {
    if (currentIndex < totalSteps) {
      onChange?.(currentIndex + 1);
      setCurrentIndex((prev) => prev + 1);
    }
  };
  const prevStep = () => {
    if (currentIndex > 1) {
      onChange?.(currentIndex - 1);
      setCurrentIndex((prev) => prev - 1);
    }
  };

  const jumpStep = (step: number) => {
    onChange?.(step);
    setCurrentIndex(step);
  };

  const reset = () => {
    onChange?.(1);
    setCurrentIndex(1);
  };

  return { currentStep: currentIndex, nextStep, prevStep, reset, jumpStep };
};

type IProgressTrackerItem = {
  active?: boolean;
  completed?: boolean;
  label: ReactNode;
  children?: ReactNode;
  onClick?: () => void;
  index: number;
  noJump?: (step: number) => boolean;
  editable?: boolean;
  step: number;
};

function ProgressTrackerItem(
  props: IProgressTrackerItem & { children?: ReactNode }
) {
  const {
    children,
    active = false,
    completed = false,
    label,
    onClick,
    index,
    noJump,
    editable,
    step,
  } = props;
  return (
    <div
      className={cn(
        'pl-5xl  border-dashed flex flex-col -mt-[10px] border-l-icon-disabled',
        {
          '[&:not(:last-child)]:border-l': !active,
          'border-l': active,
        }
      )}
    >
      <div>
        <button
          type="button"
          aria-label={`step-${index}`}
          onClick={onClick}
          className={cn(
            'border-2 border-surface-basic-default headingXs box-content w-3xl h-3xl rounded-full flex items-center justify-center absolute left-0 ',
            onClick && !noJump?.(step) ? 'cursor-pointer' : 'cursor-default',
            {
              'bg-surface-primary-default text-text-on-primary':
                !!completed && !!editable,
              'bg-icon-disabled text-text-on-primary': !!completed && !editable,
              'bg-surface-basic-active text-text-disabled':
                !completed && !active,
              'bg-surface-primary-selected text-text-default':
                active && !completed,
            }
          )}
        >
          {completed ? <Check size={12} /> : index}
        </button>
        <span
          className={cn(
            '-mt-[14px] headingMd',
            !noJump?.(step) ? 'cursor-pointer select-none' : 'cursor-default',
            {
              'text-text-default': !completed || (!!completed && !!editable),
              'text-text-disabled': !!completed && !editable,
            }
          )}
          onClick={onClick}
        >
          {label}
        </span>
      </div>
      <div className="min-h-5xl">{children}</div>
    </div>
  );
}

interface IStep {
  children?: ReactNode;
  label: ReactNode;
  step: number;
  className?: string;
  completed?: boolean;
}
const Step = ({
  children,
  step,
  className,
  label: _label,
  completed: _completed,
}: IStep) => {
  return (
    <div className={className} data-step={step}>
      {children}
    </div>
  );
};

interface IMultiStepProgress {
  children: ReactElement<IStep> | ReactElement<IStep>[];
  currentStep: number;
  jumpStep: (step: number) => void;
  noJump?: (step: number) => boolean;
  editable?: boolean;
}
const Root = ({
  children,
  currentStep,
  jumpStep,
  noJump,
  editable = true,
}: IMultiStepProgress) => {
  let child = children;
  // @ts-ignore
  if (child.type === React.Fragment) {
    // @ts-ignore
    child = child.props.children;
  }

  return (
    <div className="pl-[12px] flex flex-col relative [counter-reset:steps]">
      {Children.map(child, (ch, index) => {
        return (
          <ProgressTrackerItem
            {...ch}
            index={index + 1}
            label={ch.props.label}
            step={ch.props.step}
            active={currentStep === ch.props.step}
            noJump={noJump || (() => !(index + 1 < currentStep))}
            editable={editable}
            completed={currentStep > ch.props.step}
            onClick={() => {
              if (noJump ? !noJump?.(ch.props.step) : index + 1 < currentStep) {
                jumpStep(index + 1);
              }
            }}
          >
            {currentStep === ch.props.step ? (
              <div className="pt-3xl pb-5xl">{ch.props.children}</div>
            ) : null}
          </ProgressTrackerItem>
        );
      })}
    </div>
  );
};

const MultiStepProgress = {
  Root,
  Step,
};

export default MultiStepProgress;
