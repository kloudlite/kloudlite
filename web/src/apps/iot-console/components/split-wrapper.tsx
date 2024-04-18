import { ReactNode } from 'react';
import { ChildrenProps } from '~/components/types';

const SplitWrapper = ({
  children,
  fillerImage,
}: ChildrenProps & {
  fillerImage?: ReactNode;
}) => {
  return (
    <div className="flex">
      <div className="min-h-screen min-w-[900px] bg-surface-basic-default p-10xl flex-1 shadow-shadow-2">
        {children}
      </div>
      {fillerImage ? (
        <div className="flex mr-8xl w-[477px]">
          <div className="fixed top-0 h-screen flex items-center">
            {fillerImage}
          </div>
        </div>
      ) : (
        <div className="w-[5rem]" />
      )}
    </div>
  );
};

export default SplitWrapper;
