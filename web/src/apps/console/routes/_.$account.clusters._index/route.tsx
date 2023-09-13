import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import Wrapper from '~/console/components/wrapper';
import { listOrGrid, parseNodes } from '~/console/server/r-utils/common';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { LoadingComp, pWrapper } from '../../components/loading-component';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import { ensureAccountSet } from '../../server/utils/auth-utils';
import Resources from './resources';
import Tools from './tools';

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
            <Resources items={clusters} />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default Clusters;
