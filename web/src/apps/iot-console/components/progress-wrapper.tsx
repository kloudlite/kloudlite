import { ReactNode } from 'react';
import ProgressTracker, { IProgressTracker } from './console-progress-tracker';

type ITitleSection = {
  title: ReactNode;
  subTitle: ReactNode;
};
const TitleSection = ({ title, subTitle }: ITitleSection) => {
  return (
    <div className="flex flex-col gap-xl">
      <div className="heading4xl text-text-default">{title}</div>
      <div className="bodyLg text-text-default">{subTitle}</div>
      <div />
    </div>
  );
};

interface IProgressWrapper extends ITitleSection {
  progressItems: IProgressTracker;
  onClick?: (item: IProgressTracker['items'][number]) => void;
}
const ProgressWrapper = ({
  progressItems,
  subTitle,
  title,
  onClick,
}: IProgressWrapper) => {
  return (
    <div className="min-h-screen p-10xl pb-4xl max-w-[1024px] bg-surface-basic-default">
      <div className="max-w-[568px] flex flex-col gap-7xl">
        <TitleSection title={title} subTitle={subTitle} />
        <ProgressTracker.Root onClick={onClick} {...progressItems} />
      </div>
    </div>
  );
};

export default ProgressWrapper;
