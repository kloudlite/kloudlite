import { ReactNode } from 'react';
import { cn } from '~/components/utils';
import { Check } from '@jengaicons/react';

type IProgressTrackerItem = {
  active?: boolean;
  completed?: boolean;
  label: string;
  children?: ReactNode;
  onClick?: () => void;
  hasBorder: boolean;
  index: number;
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
  } = props;
  return (
    <div
      className={cn(
        'pl-3xl  border-dashed flex flex-col -mt-[10px]',
        { 'border-l': hasBorder || active },
        completed ? 'border-l-border-primary' : 'border-l-icon-disabled'
      )}
    >
      <div onClick={onClick}>
        <div
          className={cn(
            'border-2 border-surface-basic-default headingXs box-content w-3xl h-3xl rounded-full flex items-center justify-center absolute left-0 -ml-[12px]',
            completed
              ? 'bg-surface-primary-default text-text-on-primary'
              : 'bg-surface-primary-selected',
            active && !completed ? 'text-text-primary' : 'text-text-disabled'
          )}
        >
          {completed ? <Check size={12} /> : index}
        </div>
        <span
          className={cn(
            '-mt-[14px] headingMd',
            !active || completed ? 'text-text-disabled' : 'text-text-default'
          )}
        >
          {label}
        </span>
      </div>
      <div className="min-h-5xl">{children}</div>
    </div>
  );
}

export interface IProgressTracker {
  items: Array<Omit<IProgressTrackerItem, 'hasBorder' | 'index'>>;
  onClick?: (item: Omit<IProgressTrackerItem, 'hasBorder' | 'index'>) => void;
}
const Root = ({ items = [], onClick }: IProgressTracker) => {
  return (
    <div className="pl-[10px]">
      <div className="flex flex-col gap-y-lg relative [counter-reset:steps]">
        {items.map((item, index) => (
          <ProgressTrackerItem
            {...item}
            key={item.label}
            onClick={() => onClick?.(item)}
            hasBorder={items.length - 1 !== index}
            index={index + 1}
          />
        ))}
      </div>
    </div>
  );
};

const ProgressTracker = {
  Root,
  Item: ProgressTrackerItem,
};

export default ProgressTracker;
