import { useSearchParams } from '@remix-run/react';
import { useMemo } from 'react';
import { toast } from 'react-toastify';
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
          toast.info(`todo status`);
          return [
            { content: 'Active', value: 'active' },
            { content: 'Freezed', value: 'freezed' },
            { content: 'Archived', value: 'archived' },
          ];
        },
      },
    ],
    [searchParams]
  );

  return <CommonTools {...{ options, noViewMode: true, noSort: true }} />;
};

export default Tools;
