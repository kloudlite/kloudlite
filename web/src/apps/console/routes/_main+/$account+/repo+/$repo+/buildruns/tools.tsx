import { useSearchParams } from '@remix-run/react';
import { useMemo } from 'react';
import CommonTools from '~/console/components/common-tools';

const Tools = () => {
  const [searchParams] = useSearchParams();

  const options = useMemo(
    () => [
      {
        name: 'Status',
        type: 'status',
        search: false,
        dataFetcher: async (_: string) => {
          return [
            {
              content: 'running',
              value: true,
            },
            {
              content: 'error',
              value: false,
            },
          ];
        },
      },
    ],
    [searchParams]
  );
  return <CommonTools {...{ options }} />;
};
export default Tools;
