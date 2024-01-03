import Wrapper from '~/console/components/wrapper';
import { useLoaderData } from '@remix-run/react';
import { IRemixCtx } from '~/root/lib/types/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { defer } from '@remix-run/node';
import fake from '~/root/fake-data-generator/fake';
import SecondarySubHeader from '~/console/components/secondary-sub-header';
import Tools from './tools';
import BuildRunResources from './buildruns-resources';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  const { repo } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listBuildRuns({
      repoName: repo,
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      throw errors[0];
    }
    console.log(data);
    return { buildRunData: data };
  });

  return defer({ promise });
};

const BuildRuns = () => {
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp
      data={promise}
      skeletonData={{
        buildRunData: fake.ConsoleListNodePoolsQuery.infra_listNodePools as any,
      }}
    >
      {({ buildRunData }) => {
        const buildruns = buildRunData?.edges?.map(({ node }) => node);
        if (!buildruns) {
          return null;
        }
        const { pageInfo, totalCount } = buildRunData;
        return (
          <Wrapper
            header={{
              title: 'Build Runs',
            }}
            empty={{
              is: buildruns.length === 0,
              title: 'This is where you’ll manage your buildruns',
              content: '',
            }}
            pagination={{
              pageInfo,
              totalCount,
            }}
            tools={<Tools />}
          >
            <BuildRunResources items={buildruns} />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default BuildRuns;
