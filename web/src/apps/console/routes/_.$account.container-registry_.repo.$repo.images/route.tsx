import { defer } from '@remix-run/node';
import { useLoaderData, useParams } from '@remix-run/react';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import TagsResources from './tags-resources';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  const { repo } = ctx.params;
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(ctx.request).listDigest({
      repoName: repo,
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      tagsData: data || {},
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

const Images = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp data={promise}>
      {({ tagsData }) => {
        const tags = tagsData.edges?.map(({ node }) => node);

        return (
          <Wrapper
            empty={{
              is: tags.length === 0,
              title: 'This is where you’ll manage your projects.',
              content: (
                <p>
                  You can create a new project and manage the listed project.
                </p>
              ),
            }}
            tools={<Tools />}
          >
            <TagsResources items={tags} />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default Images;
