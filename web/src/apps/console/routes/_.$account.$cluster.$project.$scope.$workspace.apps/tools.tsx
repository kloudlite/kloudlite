import { useMemo } from 'react';
import { useSearchParams } from '@remix-run/react';
import CommonTools, { IModeProps } from '~/console/components/common-tools';
import { toast } from 'react-toastify';

const Tools = ({ viewMode, setViewMode }: IModeProps) => {
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

  return <CommonTools {...{ viewMode, setViewMode, options }} />;
};

export default Tools;
