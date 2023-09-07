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

  return <CommonTools {...{ viewMode, setViewMode, options }} />;
};
export default Tools;
