import { useState } from 'react';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Button } from '~/components/atoms/button.jsx';
import Wrapper from '~/console/components/wrapper';
import { IRemixCtx } from '~/root/lib/types/common';
import {
  getPagination,
  getSearch,
  listOrGrid,
} from '~/console/server/utils/common';
import {
  parseFromAnn,
  parseName,
  parseNodes,
} from '~/console/server/r-urils/common';
import { mapper } from '~/components/utils';
import { dayjs } from '~/components/molecule/dayjs';
import { keyconstants } from '~/console/server/r-urils/key-constants';
import ResourceList from '../../components/resource-list';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import { LoadingComp, pWrapper } from '../../components/loading-component';
import { ensureAccountSet } from '../../server/utils/auth-utils';
import Tools from './tools';

import Resources from './resources';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(ctx.request).listClusters({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });

    if (errors) {
      throw errors[0];
    }
    return {
      clustersData: data,
    };
  });

  return defer({ promise });
};

const Clusters = () => {
  const [viewMode, setViewMode] = useState<listOrGrid>('list');

  const { promise } = useLoaderData<typeof loader>();

  const { account } = useParams();

  return (
    <LoadingComp data={promise}>
      {({ clustersData }) => {
        const clusters = parseNodes(clustersData);

        if (!clusters) {
          return null;
        }

        const { pageInfo, totalCount } = clustersData;
        console.log('cluster', clusters);
        return (
          <Wrapper
            header={{
              title: 'Cluster',
              action: clusters.length > 0 && (
                <Button
                  variant="primary"
                  content="Create Cluster"
                  prefix={<PlusFill />}
                  LinkComponent={Link}
                  to={`/${account}/new-cluster`}
                />
              ),
            }}
            empty={{
              is: clusters.length === 0,
              title: 'This is where you’ll manage your cluster.',
              content: (
                <p>
                  You can create a new cluster and manage the listed cluster.
                </p>
              ),
              action: {
                content: 'Create new cluster',
                prefix: <Plus />,
                LinkComponent: Link,
                to: `/${account}/new-cluster`,
              },
            }}
            pagination={{
              pageInfo,
              totalCount,
            }}
          >
            <Tools viewMode={viewMode} setViewMode={setViewMode} />
            <Resources
              items={mapper(clusters || [], (i) => ({
                name: parseName(i),
                displayName: i.displayName,
                providerRegion:
                  `${i?.spec?.cloudProvider} (${i?.spec?.region})` || '',
                updateInfo: {
                  author: `${parseFromAnn(
                    i,
                    keyconstants.author
                  )} updated the cluster`,
                  time: dayjs(i.updateTime).fromNow(),
                },
              }))}
            />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default Clusters;
