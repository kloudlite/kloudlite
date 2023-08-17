import { useMemo } from 'react';
import { useSearchParams } from '@remix-run/react';
import CommonTools from '~/console/components/common-tools';

const Tools = ({ viewMode, setViewMode }) => {
  const [searchParams] = useSearchParams();

  const options = useMemo(
    () => [
      {
        name: 'Status',
        type: 'status',
        search: false,
        dataFetcher: async () => {
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

  return <CommonTools {...{ viewMode, setViewMode, options }} />;
};
export default Tools;
