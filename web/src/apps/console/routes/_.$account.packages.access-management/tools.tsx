import { useMemo } from 'react';
import CommonTools from '~/console/components/common-tools';

const Tools = () => {
  const options = useMemo(() => [], []);

  return <CommonTools {...{ options }} />;
};

export default Tools;
