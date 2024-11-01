import { Plus } from '~/console/components/icons';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { Button } from '@kloudlite/design-system/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { parseNodes } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import { useState } from 'react';

import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { EmptyCloudProviderImage } from '~/console/components/empty-resource-images';
import HandleProvider from './handle-provider';
import Tools from './tools';
import ProviderResourcesV2 from './provider-resources-v2';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listProviderSecrets({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      throw errors[0];
    }
    return { providersData: data };
  });

  return defer({ promise });
};

const CloudProvidersIndex = () => {
  const [visible, setVisible] = useState(false);
  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          providersData: fake.ConsoleListProviderSecretsQuery
            .infra_listProviderSecrets as any,
        }}
      >
        {({ providersData }) => {
          const providers = parseNodes(providersData);
          if (!providers) {
            return null;
          }

          const { pageInfo, totalCount } = providersData;

          return (
            <Wrapper
              secondaryHeader={{
                title: 'Cloud Provider',
                action: providers.length > 0 && (
                  <Button
                    variant="primary"
                    content="Add Cloud Provider"
                    prefix={<Plus />}
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                image: <EmptyCloudProviderImage />,
                is: providers.length === 0,
                title: 'This is where you’ll manage your cloud providers.',
                content: (
                  <p>
                    You can create a new cloud provider and manage the listed
                    cloud providers.
                  </p>
                ),
                action: {
                  content: 'Add Cloud Provider',
                  prefix: <Plus />,
                  linkComponent: Link,
                  onClick: () => {
                    setVisible(true);
                  },
                },
              }}
              pagination={{
                pageInfo,
                totalCount,
              }}
              tools={<Tools />}
            >
              <ProviderResourcesV2 items={providers} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      {/* Popup dialog for adding cloud provider */}
      <HandleProvider
        {...{
          isUpdate: false,
          visible,
          setVisible,
        }}
      />
    </>
  );
};

export default CloudProvidersIndex;
