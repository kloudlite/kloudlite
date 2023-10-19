import { defer } from '@remix-run/node';
import { useLoaderData, useParams } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import SubNavAction from '~/console/components/sub-nav-action';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { DIALOG_TYPE } from '~/console/utils/commons';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import BuildResources from './build-resources';
import HandleBuild from './handle-builds';

export const loader = async (ctx: IRemixCtx) => {
  const { repo } = ctx.params;
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(ctx.request).listBuilds({
      repoName: repo,
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      buildData: data || {},
    };
  });

  return defer({ promise });
};

const Tabs = () => {
  const { account } = useParams();
  return (
    <CommonTabs
      backButton={{
        to: `/${account}/container-registry/repos`,
        label: 'Repos',
      }}
    />
  );
};

export const handle = () => {
  return {
    navbar: <Tabs />,
  };
};

const Builds = () => {
  const [showHandleBuild, setShowHandleBuild] = useState<IShowDialog>(null);
  const { promise } = useLoaderData<typeof loader>();
  return (
    <>
      <LoadingComp data={promise}>
        {({ buildData }) => {
          const builds = buildData.edges?.map(({ node }) => node);

          return (
            <>
              <SubNavAction deps={[]}>
                <Button
                  content="Create build"
                  variant="primary"
                  onClick={() => {
                    setShowHandleBuild({ type: DIALOG_TYPE.ADD, data: null });
                  }}
                />
              </SubNavAction>
              <Wrapper
                empty={{
                  is: builds.length === 0,
                  title: 'This is where you’ll manage your projects.',
                  content: (
                    <p>
                      You can create a new project and manage the listed
                      project.
                    </p>
                  ),
                }}
                // tools={<Tools />}
              >
                <BuildResources items={builds} />
                {/* <TagsResources items={tags} /> */}
              </Wrapper>
            </>
          );
        }}
      </LoadingComp>
      <HandleBuild show={showHandleBuild} setShow={setShowHandleBuild} />
    </>
  );
};

export default Builds;
