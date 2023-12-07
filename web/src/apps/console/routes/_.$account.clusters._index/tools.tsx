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

      // {
      //   name: 'Region',
      //   type: 'region',
      //   search: false,
      //   dataFetcher: async () => {
      //     return [
      //       { content: 'Mumbai(ap-south-1)', value: 'ap-south-1' },
      //       { content: 'NY(ap-south-2)', value: 'do' },
      //     ];
      //   },
      // },

      {
        name: 'Status',
        type: 'isReady',
        search: false,
        dataFetcher: async () => {
          return [
            { content: 'Running', value: true },
            { content: 'Error', value: false },
            // { content: 'Freezed', value: false, type: 'freezed' },
          ];
        },
      },
    ],
    [searchParams]
  );

  return <CommonTools {...{ options }} />;
};

export default Tools;
