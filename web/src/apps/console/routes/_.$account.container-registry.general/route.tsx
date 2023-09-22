import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import SubNavAction from '~/console/components/sub-nav-action';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { DIALOG_DATA_NONE } from '~/console/utils/commons';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
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

const ContainerRegistryGeneral = () => {
  const [showHandleRepo, setShowHandleRepo] = useState<IShowDialog>(null);
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp data={promise}>
      {({ repository }) => {
        const repos = repository.edges?.map(({ node }) => node);
        const data = {
          action: () => {
            setShowHandleRepo(DIALOG_DATA_NONE);
          },
          content: 'Create new repository',
          show: false,
        };
        return (
          <>
            <SubNavAction data={data} visible={repos.length > 0} />
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
                    setShowHandleRepo(DIALOG_DATA_NONE);
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

export default ContainerRegistryGeneral;
