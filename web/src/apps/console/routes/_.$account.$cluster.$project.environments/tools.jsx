import { useMemo } from 'react';
import { useSearchParams, useParams } from '@remix-run/react';
import CommonTools from '~/console/components/common-tools';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';
import { parseName, parseNodes } from '~/console/server/r-urils/common';
import { isValidRegex } from '~/console/server/utils/common';

const Tools = ({ viewMode, setViewMode }) => {
  const [searchParams] = useSearchParams();

  const params = useParams();

  const api = useAPIClient();

  const options = useMemo(
    () => [
      {
        name: 'Cluster',
        type: 'clusterName',
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
              content: item.displayName,
              value: parseName(item),
            };
          });
        },
      },
    ],
    [searchParams]
  );

  return <CommonTools {...{ viewMode, setViewMode, options }} />;
};

export default Tools;
