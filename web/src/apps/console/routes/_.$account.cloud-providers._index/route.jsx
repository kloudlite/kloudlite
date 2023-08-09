import { useState } from 'react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import Wrapper from '~/console/components/Wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loader';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { toast } from '~/components/molecule/toast';
import { useReload } from '~/root/lib/client/helpers/reloader';
import ResourceList from '../../components/resource-list';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import { dummyData } from '../../dummy/data';
import {
  getPagination,
  getSearch,
  parseName,
  parseUpdationTime,
} from '../../server/r-urils/common';
import Tools from './tools';
import Filters from './filters';
import Resources from './resources';
import HandleProvider from './handle-provider';

const CloudProvidersIndex = () => {
  const [appliedFilters, setAppliedFilters] = useState(
    dummyData.appliedFilters
  );
  const [viewMode, setViewMode] = useState('list');
  const [showAddProvider, setShowAddProvider] = useState(false);
  const { promise } = useLoaderData();

  const deleteCloudProvider = async (data) => {
    console.log('delte:', parseName(data));
    toast.error('not implemented');
    // try {
    //   const { errors } = api.deleteProviderSecret({
    //     secretName: parseName(data),
    //   });
    //   if (errors) {
    //     throw errors[0];
    //   }
    //   toast.error('deleted successfully');
    //   reloadPage();
    // } catch (err) {
    //   toast.error(err.message);
    // }
  };
  return (
    <>
      <LoadingComp
        data={promise}
        skeleton={<div>Loading .....................</div>}
        errorComp={
          <div className="text-text-critical">Something went wrong</div>
        }
      >
        {({ providers, errors }) => {
          if (errors) {
            console.log(errors);
          }
          const data = providers?.edges?.map(({ node }) => node) || [];
          if (!data) {
            return null;
          }

          const { pageInfo, totalCount } = providers;
          console.log(providers);

          return (
            <Wrapper
              header={{
                title: 'Cloud Provider',
                action: data.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create Cloud Provider"
                    prefix={PlusFill}
                    onClick={() => {
                      setShowAddProvider({ type: 'add', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                is: data.length === 0,
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
                  prefix: Plus,
                  LinkComponent: Link,
                  onClick: () => {
                    setShowAddProvider({ type: 'add', data: null });
                  },
                },
              }}
              pagination={{
                pageInfo,
                totalCount,
                // currentPage,
                // itemsPerPage,
                // totalItems,
              }}
            >
              <div className="flex flex-col">
                <Tools viewMode={viewMode} setViewMode={setViewMode} />
                <Filters
                  appliedFilters={appliedFilters}
                  setAppliedFilters={setAppliedFilters}
                />
              </div>
              <ResourceList mode={viewMode}>
                {data.map((secret) => (
                  <ResourceList.ResourceItem
                    key={parseUpdationTime(secret) + parseName(secret)}
                    textValue={parseUpdationTime(secret) + parseName(secret)}
                  >
                    <Resources
                      item={secret}
                      onEdit={(e) => {
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

export const loader = async (ctx) => {
  // main promise
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
    return { providers: data };
  });

  return defer({ promise });
};

export default CloudProvidersIndex;
