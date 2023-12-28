import Wrapper from '~/console/components/wrapper';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { IRemixCtx } from '~/root/lib/types/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { defer } from '@remix-run/node';
import fake from '~/root/fake-data-generator/fake';
import Breadcrum from '~/console/components/breadcrum';
import { ChevronRight } from '@jengaicons/react';
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

const NetworkBreadcrum = () => {
  const { build } = useParams();
  return (
    <div className="flex flex-row items-center">
      <Breadcrum.Button
        content={
          <div className="flex flex-row gap-md items-center">
            <ChevronRight size={14} /> {build}
          </div>
        }
      />
    </div>
  );
};

export const handle = () => {
  return {
    navbar: <Tabs />,
    noLayout: true,
    breadcrum: () => <NetworkBreadcrum />,
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
              title: 'Buildruns',
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
