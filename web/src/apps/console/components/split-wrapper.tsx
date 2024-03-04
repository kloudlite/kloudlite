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
        <div className="flex items-center pr-8xl">{fillerImage}</div>
      ) : (
        <div className="w-[5rem]" />
      )}
    </div>
  );
};

export default SplitWrapper;
