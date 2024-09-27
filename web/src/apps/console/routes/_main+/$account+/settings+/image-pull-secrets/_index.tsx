import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { Plus, StackSimple } from '~/console/components/icons';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import HandleImagePullSecret from './handle-image-pull-secret';
import ImagePullSecretsResourcesV2 from './image-pull-secrets-resource-v2';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  // const { environment } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listImagePullSecrets({
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      throw errors[0];
    }
    return { imagePullSecretsData: data };
  });

  return defer({ promise });
};

const Routers = () => {
  const { promise } = useLoaderData<typeof loader>();
  const [visible, setVisible] = useState(false);

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          imagePullSecretsData: fake.ConsoleListImagePullSecretsQuery
            .core_listImagePullSecrets as any,
        }}
      >
        {({ imagePullSecretsData }) => {
          const imagePullSecrets = parseNodes(imagePullSecretsData);
          if (!imagePullSecrets) {
            return null;
          }
          return (
            <Wrapper
              secondaryHeader={{
                title: 'Image pull secrets',
                action: imagePullSecrets.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create image pull secret"
                    prefix={<Plus />}
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                image: <StackSimple size={48} />,
                is: imagePullSecrets.length === 0,
                title: 'This is where you’ll manage your image pull secrets.',
                content: (
                  <p>
                    You can create a new image pull secret and manage the listed
                    image pull secrets.
                  </p>
                ),
                action: {
                  content: 'Add new image pull secret',
                  prefix: <Plus />,
                  onClick: () => {
                    setVisible(true);
                  },
                },
              }}
              tools={<Tools />}
            >
              <ImagePullSecretsResourcesV2 items={imagePullSecrets} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleImagePullSecret {...{ visible, setVisible, isUpdate: false }} />
    </>
  );
};

export default Routers;
