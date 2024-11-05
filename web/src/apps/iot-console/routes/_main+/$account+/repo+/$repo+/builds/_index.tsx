import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import {
  LoadingComp,
  pWrapper,
} from '~/iotconsole/components/loading-component';
import Wrapper from '~/iotconsole/components/wrapper';
import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import { ensureAccountSet } from '~/iotconsole/server/utils/auth-utils';
import { getPagination, getSearch } from '~/iotconsole/server/utils/common';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import { Plus } from '~/iotconsole/components/icons';
import { Button } from '@kloudlite/design-system/atoms/button';
import fake from '~/root/fake-data-generator/fake';
import BuildResources from './build-resources';
import HandleBuild from './handle-builds';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  const { repo } = ctx.params;
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(ctx.request).listBuilds({
      repoName: atob(repo),
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

const Builds = () => {
  const [visible, setVisible] = useState(false);
  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          buildData: fake.ConsoleListBuildsQuery.cr_listBuilds as any,
        }}
      >
        {({ buildData }) => {
          const builds = buildData.edges?.map(({ node }) => node);

          return (
            <Wrapper
              header={{
                title: 'Build Integrations',
                action: builds.length > 0 && (
                  <Button
                    content="Create build"
                    variant="primary"
                    to="../new-build"
                    linkComponent={Link}
                    prefix={<Plus />}
                  />
                ),
              }}
              empty={{
                is: builds.length === 0,
                title: 'This is where you’ll manage your Build Integrations.',
                action: {
                  content: 'create build',

                  to: '../new-build',
                  linkComponent: Link,
                  prefix: <Plus />,
                },
                content: (
                  <p>
                    You can create a new Build Integration and manage the listed
                    Build Integrations.
                  </p>
                ),
              }}
              tools={<Tools />}
            >
              <BuildResources items={builds} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleBuild {...{ isUpdate: false, visible, setVisible }} />
    </>
  );
};

export default Builds;
