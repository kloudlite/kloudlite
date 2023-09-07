import { useMemo } from 'react';
import { useSearchParams } from '@remix-run/react';
import CommonTools from '~/console/components/common-tools';
import { IToolsProps } from '~/console/server/utils/common';

const Tools = ({ viewMode, setViewMode }: IToolsProps) => {
  const [searchParams] = useSearchParams();

  const options = useMemo(
    () => [
      {
        name: 'Provider',
        type: 'cloudProviderName',
        search: false,
        dataFetcher: async () => {
          return [
            { content: 'Amazon Web Services', value: 'aws' },
            { content: 'Digital Ocean', value: 'do' },
            { content: 'Google Cloud Platform', value: 'gcp' },
            { content: 'Microsoft Azure', value: 'azure' },
          ];
        },
      },
    ],
    [searchParams]
  );

  return <CommonTools {...{ viewMode, setViewMode, options }} />;
};

export default Tools;
