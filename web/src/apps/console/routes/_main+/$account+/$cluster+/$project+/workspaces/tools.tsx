import { useSearchParams } from '@remix-run/react';
import { useMemo } from 'react';
import CommonTools from '~/console/components/common-tools';

const Tools = () => {
  const [searchParams] = useSearchParams();

  const options = useMemo(() => [], [searchParams]);

  return <CommonTools {...{ options }} />;
};

export default Tools;
