import { Link } from '@remix-run/react';
import { ReactNode } from 'react';
import { Button } from '~/components/atoms/button';
import { ArrowLeft } from '~/iotconsole/components/icons';
import { cn } from '~/components/utils';
import SplitWrapper from './split-wrapper';

type ITitleSection = {
  title: ReactNode;
  subTitle: ReactNode;
  action?: ReactNode;
};
const TitleSection = ({ title, subTitle, action }: ITitleSection) => {
  return (
    <div className="flex flex-col gap-xl">
      <div className="flex flex-row justify-between">
        <div className="heading4xl text-text-default">{title}</div>
        {action}
      </div>
      <div className="bodyLg text-text-default">{subTitle}</div>
      <div />
    </div>
  );
};

interface IProgressWrapper extends ITitleSection {
  children: ReactNode;
  backButton?: {
    to: string;
    content: string;
  };
  fillerImage?: ReactNode;
  className?: string;
}
const MultiStepProgressWrapper = ({
  subTitle,
  title,
  children,
  backButton,
  fillerImage,
  action,
  className,
}: IProgressWrapper) => {
  return (
    <SplitWrapper fillerImage={fillerImage}>
      <div className={cn('max-w-[568px] flex flex-col gap-7xl', className)}>
        <div className="flex flex-col gap-xl">
          {backButton && (
            <Button
              variant="plain"
              prefix={<ArrowLeft />}
              size="sm"
              content={backButton.content}
              to={backButton.to}
              linkComponent={Link}
            />
          )}
          <TitleSection title={title} subTitle={subTitle} action={action} />
        </div>
        {children}
      </div>
    </SplitWrapper>
  );
};

export default MultiStepProgressWrapper;
