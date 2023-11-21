import { ReactNode, useEffect } from 'react';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';

const SubNavAction = ({
  children,
  deps,
}: {
  children: ReactNode;
  deps: Array<any>;
}) => {
  const subNavAction = useSubNavData();

  useEffect(() => {
    subNavAction.setData(children);
    console.log('helre', children);
  }, deps);
  return null;
};

export default SubNavAction;
