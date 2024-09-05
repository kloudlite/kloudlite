import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { Dockerlogo } from '~/console/components/icons';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination } from '~/console/server/utils/common';
import { IRemixCtx } from '~/lib/types/common';
import ImagesResource from './images-resources';
import Tools from './tools';

export const loader = (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);

    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listRegistryImages({
      pq: getPagination(ctx),
    });

    if (errors) {
      throw errors[0];
    }
    return { imagesData: data };
  });
  return defer({ promise });
};

const Images = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp
      data={promise}
    // skeletonData={{
    //   imagesData: fake.ConsoleListRegistryImagesQuery.core_listRegistryImages
    //     as any,

    // }}
    >
      {({ imagesData }) => {
        const images = parseNodes(imagesData);

        return (
          <Wrapper
            secondaryHeader={{
              title: 'Images',
            }}
            empty={{
              image: <Dockerlogo size={48} />,
              is: images.length === 0,
              title: 'This is where you’ll manage your registry images.',
              content: <p>You will get all your registry images here.</p>,
              // action: {
              //   content: 'Create new managed resource',
              //   prefix: <Plus />,
              //   to: '../new-managed-resource',
              //   linkComponent: Link,
              // },
            }}
            tools={<Tools />}
            pagination={imagesData}
          >
            <ImagesResource items={images} />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default Images;
