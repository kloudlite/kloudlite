import { useMemo } from 'react';
import { useSearchParams, useParams } from '@remix-run/react';
import CommonTools, {
  ICommonToolsOption,
} from '~/console/components/common-tools';
import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';
import { isValidRegex } from '~/console/server/utils/common';
import { parseName, parseNodes } from '~/console/server/r-utils/common';
import { useConsoleApi } from '~/console/server/gql/api-provider';

const Tools = ({ viewMode, setViewMode }: any) => {
  const [searchParams] = useSearchParams();

  const params = useParams();

  const api = useConsoleApi();

  const options: ICommonToolsOption[] = useMemo(
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
