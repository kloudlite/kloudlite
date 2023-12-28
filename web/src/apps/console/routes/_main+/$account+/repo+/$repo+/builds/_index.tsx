import { defer } from '@remix-run/node';
import { useLoaderData, useParams } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import SubNavAction from '~/console/components/sub-nav-action';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import SecondarySubHeader from '~/console/components/secondary-sub-header';
import BuildResources from './build-resources';
import HandleBuild from './handle-builds';
import Tools from './tools';

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
  const [visible, setVisible] = useState(false);
  const { promise } = useLoaderData<typeof loader>();
  return (
    <>
      <LoadingComp data={promise}>
        {({ buildData }) => {
          const builds = buildData.edges?.map(({ node }) => node);

          return (
            <div className="flex flex-col gap-3xl">
              <SecondarySubHeader
                title="Builds"
                action={
                  <Button
                    content="Create build"
                    variant="primary"
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                }
              />
              <Wrapper
                empty={{
                  is: builds.length === 0,
                  title: 'This is where you’ll manage your Build Configs.',
                  content: (
                    <p>
                      You can create a new Build Config and manage the listed
                      Build Configs.
                    </p>
                  ),
                }}
                tools={<Tools />}
              >
                <BuildResources items={builds} />
              </Wrapper>
            </div>
          );
        }}
      </LoadingComp>
      <HandleBuild {...{ isUpdate: false, visible, setVisible }} />
    </>
  );
};

export default Builds;
