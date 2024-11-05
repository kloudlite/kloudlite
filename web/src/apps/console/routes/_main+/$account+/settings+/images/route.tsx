import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '@kloudlite/design-system/atoms/button';
import { Dockerlogo, Info } from '~/console/components/icons';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination } from '~/console/server/utils/common';
import { IRemixCtx } from '~/lib/types/common';
import { RegistryImageInstruction } from './handle-image-discovery';
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
  const [visible, setVisible] = useState(false);
  return (
    <>
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
                action: images.length > 0 && (
                  <Button
                    variant="outline"
                    size="md"
                    content="Add your images"
                    prefix={<Info />}
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                image: <Dockerlogo size={48} />,
                is: images.length === 0,
                title: 'This is where you’ll manage your registry images.',
                content: <p>You will get all your registry images here.</p>,
                action: {
                  content: 'Add your images',
                  size: 'md',
                  variant: 'outline',
                  prefix: <Info />,
                  onClick: () => {
                    setVisible(true);
                  },
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
      <RegistryImageInstruction
        {...{ show: visible, onClose: () => setVisible(false) }}
      />
    </>
  );
};

export default Images;
