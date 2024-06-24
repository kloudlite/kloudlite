import { Plus } from '~/console/components/icons';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import HandleScope from '~/console/page-components/handle-environment';
import { parseNodes } from '~/console/server/r-utils/common';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';

import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { IEnvironment } from '~/console/server/gql/queries/environment-queries';
import { EmptyEnvironmentImage } from '~/console/components/empty-resource-images';
import Tools from './tools';
import EnvironmentResourcesV2 from './environment-resources-v2';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listEnvironments({
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });

    if (errors) {
      throw errors[0];
    }

    return {
      environmentData: data || {},
    };
  });

  return defer({ promise });
};

const Workspaces = () => {
  const [showAddWS, setShowAddWS] =
    useState<IShowDialog<IEnvironment | null>>(null);

  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          environmentData: fake.ConsoleListEnvironmentsQuery
            .core_listEnvironments as any,
        }}
      >
        {({ environmentData }) => {
          const environments = parseNodes(environmentData);

          if (!environments) {
            return null;
          }

          return (
            <Wrapper
              header={{
                title: 'Environments',
                action: environments.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create Environment"
                    prefix={<Plus />}
                    onClick={() => {
                      setShowAddWS({ type: 'add', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                image: <EmptyEnvironmentImage />,
                is: environments?.length === 0,
                title: 'This is where you’ll manage your environment.',
                content: (
                  <p>
                    You can create a new workspace and manage the listed
                    workspaces.
                  </p>
                ),
                action: {
                  content: 'Create new environment',
                  prefix: <Plus />,
                  onClick: () => {
                    setShowAddWS({ type: 'add', data: null });
                  },
                },
              }}
              tools={<Tools />}
            >
              <EnvironmentResourcesV2 items={environments || []} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleScope show={showAddWS} setShow={setShowAddWS} />
    </>
  );
};
export default Workspaces;
