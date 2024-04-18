import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { LoadingComp, pWrapper } from '~/iotconsole/components/loading-component';
import Wrapper from '~/iotconsole/components/wrapper';
import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import { ensureAccountSet } from '~/iotconsole/server/utils/auth-utils';
import { getPagination, getSearch } from '~/iotconsole/server/utils/common';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import TagsResources from './tags-resources';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  const { repo } = ctx.params;

  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(ctx.request).listDigest({
      repoName: atob(repo),
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

const Images = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp
      data={promise}
      skeletonData={{
        tagsData: fake.ConsoleListDigestQuery.cr_listDigests as any,
      }}
    >
      {({ tagsData }) => {
        const tags = tagsData.edges?.map(({ node }) => node);

        return (
          <Wrapper
            header={{
              title: 'Images',
            }}
            empty={{
              is: tags.length === 0,
              title: 'This is where you’ll manage your images.',
              content: (
                <p>
                  You can push images to this repository and start using them in
                  your deployments.
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
