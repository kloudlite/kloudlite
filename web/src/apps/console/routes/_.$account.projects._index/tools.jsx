import { useMemo } from 'react';
import { useSearchParams, useParams } from '@remix-run/react';
import CommonTools from '~/console/components/common-tools';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';
import {
  isValidRegex,
  parseDisplaynameFromAnn,
  parseName,
} from '~/console/server/r-urils/common';
import { toast } from 'react-toastify';
import { parseNodes } from '~/root/src/generated/r-types/utils';

const Tools = ({ viewMode, setViewMode }) => {
  const [searchParams] = useSearchParams();

  const params = useParams();

  const api = useAPIClient();

  const options = useMemo(
    () => [
      {
        name: 'Cluster',
        type: 'text',
        search: true,
        dataFetcher: async (s) => {
          ensureAccountClientSide(params);
          const { data, errors } = await api.listClusters(
            isValidRegex(s)
              ? {
                  search: {
                    text: {
                      matchType: 'regex',
                      regex: s || '',
                    },
                  },
                }
              : {}
          );

          if (errors) {
            throw errors[0];
          }

          const datas = parseNodes(data);
          return datas.map((item) => {
            return {
              content: parseDisplaynameFromAnn(item),
              value: parseName(item),
            };
          });
        },
      },
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
