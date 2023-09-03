import { useMemo } from 'react';
import CommonTools from '~/console/components/common-tools';

const Tools = () => {
  const options = useMemo(() => [], []);

  return <CommonTools {...{ options, noViewMode: true }} />;
};

export default Tools;
