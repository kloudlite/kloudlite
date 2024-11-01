import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { Button } from '@kloudlite/design-system/atoms/button';
import { EmptyManagedResourceImage } from '~/console/components/empty-resource-images';
import { Plus } from '~/console/components/icons';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/lib/types/common';
import ImagesResource from './images-resources';
import Tools from './tools';

export const loader = (ctx: IRemixCtx) => {
  const { imagepullsecret } = ctx.params;
  console.log('====>>>>', imagepullsecret);
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);

    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listEnvironments({
      pq: getPagination(ctx),
      search: getSearch(ctx),
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
      //   skeletonData={{
      //     managedResourcesData: fake.ConsoleListManagedResourcesQuery
      //       .core_listManagedResources as any,
      //   }}
    >
      {({ imagesData }) => {
        const images = parseNodes(imagesData);

        return (
          <Wrapper
            header={{
              title: 'Images',
              action: images.length > 0 && (
                <Button
                  variant="primary"
                  content="Create managed resource"
                  prefix={<Plus />}
                  to="../new-managed-resource"
                  linkComponent={Link}
                />
              ),
            }}
            empty={{
              image: <EmptyManagedResourceImage />,
              is: images.length === 0,
              title: 'This is where you’ll manage your managed resources.',
              content: (
                <p>
                  You can create a new managed resource and manage the listed
                  Managed resource.
                </p>
              ),
              action: {
                content: 'Create new managed resource',
                prefix: <Plus />,
                to: '../new-managed-resource',
                linkComponent: Link,
              },
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
