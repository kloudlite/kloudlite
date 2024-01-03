import { defer } from '@remix-run/node';
import { Link, useLoaderData, useParams } from '@remix-run/react';
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
import { Plus } from '@jengaicons/react';
import SecondarySubHeader from '~/console/components/secondary-sub-header';
import Tools from './tools';
import BuildCachesResources from './build-caches-resources';
import HandleBuildCache from './handle-build-cache';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listBuildCaches({
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      buildCachesData: data || {},
    };
  });

  return defer({ promise });
};

const Builds = () => {
  const [visible, setVisible] = useState(false);
  const { promise } = useLoaderData<typeof loader>();
  return (
    <>
      <LoadingComp data={promise}>
        {({ buildCachesData }) => {
          const buildsCaches = buildCachesData.edges?.map(({ node }) => node);

          return (
            <div className="flex flex-col gap-6xl">
              <SecondarySubHeader
                title="Build caches"
                action={
                  buildsCaches.length > 0 && (
                    <Button
                      content="Create build cache"
                      variant="primary"
                      onClick={() => {
                        setVisible(true);
                      }}
                    />
                  )
                }
              />
              <Wrapper
                empty={{
                  is: buildsCaches.length === 0,
                  title: 'This is where you’ll manage your projects.',
                  content: (
                    <p>
                      You can create a new project and manage the listed
                      project.
                    </p>
                  ),
                  action: {
                    content: 'Create build cache',
                    prefix: <Plus />,
                    LinkComponent: Link,
                    onClick: () => {
                      setVisible(true);
                    },
                  },
                }}
                tools={<Tools />}
              >
                <BuildCachesResources items={buildsCaches} />
              </Wrapper>
            </div>
          );
        }}
      </LoadingComp>
      <HandleBuildCache {...{ isUpdate: false, visible, setVisible }} />
    </>
  );
};

export default Builds;
