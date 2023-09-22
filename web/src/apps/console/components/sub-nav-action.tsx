import { useEffect } from 'react';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import { ISubNavCallback } from './types.d';

const SubNavAction = ({
  data,
  visible = false,
}: {
  data: ISubNavCallback;
  visible?: boolean;
}) => {
  const subNavAction = useSubNavData();
  const tempData = data;
  if (!visible) {
    tempData.show = false;
  } else {
    tempData.show = true;
  }

  useEffect(() => {
    subNavAction.setData(tempData);
  }, [visible, data]);
  return null;
};

export default SubNavAction;
