import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import { IProviderSecrets } from '~/console/server/gql/queries/provider-secret-queries';
import { ExtractNodeType, parseNodes } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { DIALOG_TYPE } from '~/console/utils/commons';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import { GQLServerHandler } from '../../server/gql/saved-queries';

import HandleProvider from './handle-provider';
import ProviderResources from './provider-resources';
import Tools from './tools';

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
                    prefix={<PlusFill />}
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                is: providers.length === 0,
                title: 'you have not added any cloud provider yet.',
                content: (
                  <p>
                    please add some cloud providers to start creating cluster.
                  </p>
                ),
                action: {
                  content: 'Add Cloud Provider',
                  prefix: <Plus />,
                  LinkComponent: Link,
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
              <ProviderResources items={providers} />
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
