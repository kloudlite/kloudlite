import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import { getPagination, getSearch } from '~/console/server/utils/common';
import Tools from './tools';
import StorageResources from './storage-resources';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { cluster } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listPvcs({
      clusterName: cluster,
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      throw errors[0];
    }
    return { storageData: data };
  });

  return defer({ promise });
};

const ClusterStorage = () => {
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp
      data={promise}
      skeletonData={{
        storageData: fake.ConsoleListNodePoolsQuery.infra_listNodePools as any,
      }}
    >
      {({ storageData }) => {
        const storages = storageData?.edges?.map(({ node }) => node);
        if (!storages) {
          return null;
        }
        const { pageInfo, totalCount } = storageData;
        return (
          <Wrapper
            header={{
              title: 'Storage',
            }}
            empty={{
              is: storages.length === 0,
              title: 'This is where you’ll manage your storage',
              content: '',
            }}
            pagination={{
              pageInfo,
              totalCount,
            }}
            tools={<Tools />}
          >
            <StorageResources items={storages} />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default ClusterStorage;
