import { ChildrenProps } from '~/components/types';

const OnlyLatest = ({
  children,
}: ChildrenProps & {
  last: string;
}) => {
  return children;
};

export default OnlyLatest;
