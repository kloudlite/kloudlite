import { useState } from 'react';
import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import Wrapper from '~/console/components/wrapper';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { toast } from '~/components/molecule/toast';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { Button } from '~/components/atoms/button';
import { parseName, parseNodes } from '~/console/server/r-urils/common';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { IProviderSecret } from '~/console/server/gql/queries/provider-secret-queries';
import ResourceList from '../../components/resource-list';
import { GQLServerHandler } from '../../server/gql/saved-queries';

import Tools from './tools';
import Resources from './resources';
import HandleProvider from './handle-provider';

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
  const [viewMode, setViewMode] = useState<'list' | 'grid'>('list');
  const [showAddProvider, setShowAddProvider] = useState<any>(null);
  const { promise } = useLoaderData<typeof loader>();

  const deleteCloudProvider = async (data: IProviderSecret) => {
    console.log('delete:', parseName(data));
    toast.error('not implemented');
  };

  return (
    <>
      <LoadingComp data={promise}>
        {({ providersData }) => {
          const providers = parseNodes(providersData);
          if (!providers) {
            return null;
          }

          const { pageInfo, totalCount } = providersData;

          return (
            <Wrapper
              header={{
                title: 'Cloud Provider',
                action: providers.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create Cloud Provider"
                    prefix={<PlusFill />}
                    onClick={() => {
                      setShowAddProvider({ type: 'add', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                is: providers.length === 0,
                title:
                  'This is the place where you will oversees the Cloud Provider.',
                content: (
                  <p>
                    You have the option to include a new Cloud Provider and
                    oversee the existing Cloud Provider.
                  </p>
                ),
                action: {
                  content: 'Create Cloud Provider',
                  prefix: <Plus />,
                  LinkComponent: Link,
                  onClick: () => {
                    setShowAddProvider({ type: 'add', data: null });
                  },
                },
              }}
              pagination={{
                pageInfo,
                totalCount,
              }}
            >
              <Tools viewMode={viewMode} setViewMode={setViewMode} />
              <ResourceList mode={viewMode}>
                {providers.map((secret) => (
                  <ResourceList.ResourceItem
                    key={secret.updateTime + parseName(secret)}
                    textValue={secret.updateTime + parseName(secret)}
                  >
                    <Resources
                      item={secret}
                      onEdit={(e: any) => {
                        setShowAddProvider({ type: 'edit', data: e });
                      }}
                      onDelete={deleteCloudProvider}
                    />
                  </ResourceList.ResourceItem>
                ))}
              </ResourceList>
            </Wrapper>
          );
        }}
      </LoadingComp>
      {/* Popup dialog for adding cloud provider */}
      <HandleProvider show={showAddProvider} setShow={setShowAddProvider} />
    </>
  );
};

export default CloudProvidersIndex;
