import { useSearchParams } from '@remix-run/react';
import { useMemo } from 'react';
import CommonTools from '~/console/components/common-tools';

const Tools = () => {
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

  return <CommonTools {...{ options }} />;
};

export default Tools;
