import { ReactNode } from 'react';

interface ISecondarySubHeader {
  title: ReactNode;
  action: ReactNode;
}
const SecondarySubHeader = ({ title, action }: ISecondarySubHeader) => {
  return (
    <div className="flex flex-row items-center min-h-[36px]">
      <div className="headingXl text-text-strong flex-1">{title}</div>
      <div>{action}</div>
    </div>
  );
};

export default SecondarySubHeader;
