import { Check } from '@jengaicons/react';
import React, {
  Children,
  ReactElement,
  ReactNode,
  useRef,
  useState,
} from 'react';
import { cn } from '~/components/utils';

interface IUseMultiStepProgress {
  defaultStep: number;
  totalSteps: number;
}
export const useMultiStepProgress = ({
  defaultStep = 1,
  totalSteps = 1,
}: IUseMultiStepProgress) => {
  const [currentIndex, setCurrentIndex] = useState(defaultStep);

  const nextStep = () => {
    if (currentIndex < totalSteps) {
      setCurrentIndex((prev) => prev + 1);
    }
  };
  const prevStep = () => {
    if (currentIndex > 1) {
      setCurrentIndex((prev) => prev - 1);
    }
  };

  const jumpStep = (step: number) => {
    setCurrentIndex(step);
  };

  const reset = () => {
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
  hasBorder: boolean;
  index: number;
  noJump?: boolean;
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
    hasBorder,
    index,
    noJump,
  } = props;

  return (
    <div
      className={cn(
        'pl-3xl  border-dashed flex flex-col -mt-[10px]',
        { 'border-l': hasBorder || active },
        completed ? 'border-l-border-primary' : 'border-l-icon-disabled'
      )}
    >
      <div>
        <button
          type="button"
          aria-label={`step-${index}`}
          onClick={onClick}
          className={cn(
            'border-2 border-surface-basic-default headingXs box-content w-3xl h-3xl rounded-full flex items-center justify-center absolute left-0 -ml-[12px]',
            completed
              ? 'bg-surface-primary-default text-text-on-primary'
              : 'bg-surface-primary-selected',
            active && !completed ? 'text-text-primary' : 'text-text-disabled',
            onClick && !noJump ? 'cursor-pointer' : 'cursor-default'
          )}
        >
          {completed ? <Check size={12} /> : index}
        </button>
        <span
          className={cn(
            '-mt-[14px] headingMd',
            !noJump ? 'cursor-pointer select-none' : '',
            !active || completed ? 'text-text-disabled' : 'text-text-default'
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
}
const Step = ({ children, step, className, label: _label }: IStep) => {
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
  noJump?: boolean;
}
const Root = ({
  children,
  currentStep,
  jumpStep,
  noJump,
}: IMultiStepProgress) => {
  let child = children;
  // @ts-ignore
  if (child.type === React.Fragment) {
    // @ts-ignore
    child = child.props.children;
  }

  return (
    <div className="flex flex-col relative [counter-reset:steps]">
      {Children.map(child, (ch, index) => {
        return (
          <ProgressTrackerItem
            {...ch}
            index={index + 1}
            label={ch.props.label}
            hasBorder={ch.props.step <= currentStep}
            active={currentStep === ch.props.step}
            noJump={noJump}
            onClick={() => {
              if (index + 1 < currentStep) {
                if (!noJump) {
                  jumpStep(index + 1);
                }
              }
            }}
          >
            {currentStep === ch.props.step ? (
              <div className="py-3xl">{ch.props.children}</div>
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
