import { Link } from '@remix-run/react';
import { ReactNode } from 'react';
import { Button } from '~/components/atoms/button';
import { ArrowLeft } from '~/console/components/icons';

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
  children: ReactNode;
  backButton?: {
    to: string;
    content: string;
  };
}
const MultiStepProgressWrapper = ({
  subTitle,
  title,
  children,
  backButton,
}: IProgressWrapper) => {
  return (
    <div className="min-h-screen p-10xl pb-4xl max-w-[1024px] bg-surface-basic-default">
      <div className="max-w-[568px] flex flex-col gap-7xl">
        <div className="flex flex-col gap-xl">
          {backButton && (
            <Button
              variant="plain"
              prefix={<ArrowLeft />}
              size="sm"
              content={backButton.content}
              to={backButton.to}
              LinkComponent={Link}
            />
          )}
          <TitleSection title={title} subTitle={subTitle} />
        </div>
        {children}
      </div>
    </div>
  );
};

export default MultiStepProgressWrapper;
