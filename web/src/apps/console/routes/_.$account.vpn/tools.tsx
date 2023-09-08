import { useMemo } from 'react';
import { useSearchParams } from '@remix-run/react';
import CommonTools from '~/console/components/common-tools';
import { IToolsProps } from '~/console/server/utils/common';

const Tools = ({ viewMode, setViewMode }: IToolsProps) => {
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
