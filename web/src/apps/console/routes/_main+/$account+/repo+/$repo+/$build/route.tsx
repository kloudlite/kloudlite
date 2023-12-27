import Wrapper from '~/console/components/wrapper';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { useLoaderData, useParams } from '@remix-run/react';
import { IRemixCtx } from '~/root/lib/types/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { defer } from '@remix-run/node';
import fake from '~/root/fake-data-generator/fake';
import Tools from './tools';

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
    return { buildRunData: data };
  });

  return defer({ promise });
};

const Tabs = () => {
  const { account } = useParams();
  return (
    <CommonTabs
      backButton={{
        to: `/${account}/packages/`,
        label: 'Build configs',
      }}
    />
  );
};

export const handle = () => {
  return {
    navbar: <Tabs />,
  };
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
              title: 'Storage',
            }}
            empty={{
              is: buildruns.length === 0,
              title: 'This is where you’ll manage your storage',
              content: '',
            }}
            pagination={{
              pageInfo,
              totalCount,
            }}
            tools={<Tools />}
          >
            {/* <BuildRunResources items={buildruns} /> */}
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default BuildRuns;
