import { useSearchParams } from '@remix-run/react';
import { useMemo } from 'react';
import CommonTools from '~/console/components/common-tools';

const Tools = () => {
  const [searchParams] = useSearchParams();

  const options = useMemo(
    () => [
      {
        name: 'Status',
        type: 'text',
        search: false,
        dataFetcher: async () => {
          return [
            { content: 'Running', value: 'running' },
            { content: 'Warning', value: 'warning' },
            { content: 'Freezed', value: 'freezed' },
            { content: 'Error', value: 'error' },
          ];
        },
      },
    ],
    [searchParams]
  );

  return <CommonTools {...{ options }} />;
};

export default Tools;
