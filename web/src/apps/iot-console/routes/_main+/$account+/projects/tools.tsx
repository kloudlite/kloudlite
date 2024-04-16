import { useSearchParams } from '@remix-run/react';
import { useMemo } from 'react';
import CommonTools from '~/iotconsole/components/common-tools';

const Tools = () => {
  const [searchParams] = useSearchParams();

  // const options: FilterType[] = useMemo(
  //   () => [
  //     {
  //       name: 'Cluster',
  //       type: 'text',
  //       search: true,
  //       dataFetcher: async (s) => {
  //         ensureAccountClientSide(params);
  //         const { data, errors } = await api.listClusters(
  //           isValidRegex(s)
  //             ? {
  //                 search: {
  //                   text: {
  //                     matchType: 'regex',
  //                     regex: s || '',
  //                   },
  //                 },
  //               }
  //             : {}
  //         );

  //         if (errors) {
  //           throw errors[0];
  //         }

  //         const datas = parseNodes(data);
  //         return datas.map((item) => {
  //           return {
  //             content: item.displayName,
  //             value: parseName(item),
  //           };
  //         });
  //       },
  //     },
  //   ],
  //   [searchParams]
  // );
  const options = useMemo(() => [], [searchParams]);

  return <CommonTools {...{ options }} />;
};

export default Tools;
