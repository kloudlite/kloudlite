import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
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
import { MainLayoutSK } from '~/console/page-components/skeletons';
import HandleRepo from './handle-repo';
import RepoResources from './repo-resources';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(ctx.request).listRepo({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      repository: data || {},
    };
  });

  return defer({ promise });
};

const ContainerRegistryRepos = () => {
  const [showHandleRepo, setShowHandleRepo] = useState<IShowDialog>(null);
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp
      data={promise}
      skeleton={<MainLayoutSK title="Repos" noHeader />}
    >
      {({ repository }) => {
        const repos = repository.edges?.map(({ node }) => node);

        return (
          <>
            {repos.length > 0 && (
              <SubNavAction deps={[repos.length]}>
                <Button
                  content="Create new repository"
                  variant="primary"
                  onClick={() => {
                    setShowHandleRepo({ type: DIALOG_TYPE.ADD, data: null });
                  }}
                />
              </SubNavAction>
            )}
            <Wrapper
              empty={{
                is: repos?.length === 0,
                title: 'This is where you’ll manage your repository.',
                content: (
                  <p>
                    You can create a new repository and manage the listed
                    repository.
                  </p>
                ),
                action: {
                  content: 'Create repository',
                  prefix: <Plus />,
                  onClick: () => {
                    setShowHandleRepo({ type: DIALOG_TYPE.ADD, data: null });
                  },
                },
              }}
              tools={<Tools />}
            >
              <RepoResources items={repos} />
            </Wrapper>
            <HandleRepo show={showHandleRepo} setShow={setShowHandleRepo} />
          </>
        );
      }}
    </LoadingComp>
  );
};

export default ContainerRegistryRepos;
