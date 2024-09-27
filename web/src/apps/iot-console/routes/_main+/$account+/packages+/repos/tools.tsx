import { useMemo } from 'react';
import CommonTools from '~/iotconsole/components/common-tools';

const Tools = () => {
  const options = useMemo(() => [], []);

  return <CommonTools {...{ options }} />;
};

export default Tools;
