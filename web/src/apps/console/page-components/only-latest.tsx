import { ChildrenProps } from '@kloudlite/design-system/types';

const OnlyLatest = ({
  children,
}: ChildrenProps & {
  last: string;
}) => {
  return children;
};

export default OnlyLatest;
