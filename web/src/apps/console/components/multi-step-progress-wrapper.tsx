import { Link } from '@remix-run/react';
import { ReactNode } from 'react';
import { Button } from '~/components/atoms/button';
import { ArrowLeft } from '~/console/components/icons';
import SplitWrapper from './split-wrapper';

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
  fillerImage?: ReactNode;
}
const MultiStepProgressWrapper = ({
  subTitle,
  title,
  children,
  backButton,
  fillerImage,
}: IProgressWrapper) => {
  return (
    <SplitWrapper fillerImage={fillerImage}>
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
    </SplitWrapper>
  );
};

export default MultiStepProgressWrapper;
