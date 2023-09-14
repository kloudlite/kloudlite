import { useParams, useSearchParams } from '@remix-run/react';
import { useMemo } from 'react';
import CommonTools from '~/console/components/common-tools';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName, parseNodes } from '~/console/server/r-utils/common';
import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';
import { isValidRegex } from '~/console/server/utils/common';

const Tools = () => {
  const [searchParams] = useSearchParams();

  const params = useParams();

  const api = useConsoleApi();

  const options = useMemo(
    () => [
      {
        name: 'Cluster',
        type: 'clusterName',
        search: true,
        dataFetcher: async (s: string) => {
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
              content: item.displayName,
              value: parseName(item),
            };
          });
        },
      },
    ],
    [searchParams]
  );

  return <CommonTools {...{ options }} />;
};

export default Tools;
