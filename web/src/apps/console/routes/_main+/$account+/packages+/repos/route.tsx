import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import SecondarySubHeader from '~/console/components/secondary-sub-header';
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
  const [visible, setVisible] = useState(false);
  const { promise } = useLoaderData<typeof loader>();
  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          repository: fake.ConsoleListRepoQuery.cr_listRepos as any,
        }}
      >
        {({ repository }) => {
          const repos = repository.edges?.map(({ node }) => node);

          return (
            <div className="flex flex-col gap-6xl">
              <SecondarySubHeader
                title={
                  <div className="flex flex-row gap-xl items-center">
                    <span>Container Repos</span>
                  </div>
                }
                action={
                  <Button
                    content="Create new repository"
                    variant="primary"
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                }
              />
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
                      setVisible(true);
                    },
                  },
                }}
                tools={<Tools />}
              >
                <RepoResources items={repos} />
              </Wrapper>
            </div>
          );
        }}
      </LoadingComp>
      <HandleRepo
        {...{
          isUpdate: false,
          visible,
          setVisible,
        }}
      />
    </>
  );
};

export default ContainerRegistryRepos;
