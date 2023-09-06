import { useMemo } from 'react';
import { useSearchParams } from '@remix-run/react';
import CommonTools from '~/console/components/common-tools';

// @ts-ignore
const Tools = ({ viewMode, setViewMode }) => {
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

  return <CommonTools {...{ viewMode, setViewMode, options }} />;
};

export default Tools;
