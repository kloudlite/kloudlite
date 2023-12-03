import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData, useOutletContext } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import {
  getScopeAndProjectQuery,
  parseNodes,
} from '~/console/server/r-utils/common';
import { getSearch } from '~/console/server/utils/common';
import { getManagedTemplate } from '~/console/utils/commons';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '../../server/utils/auth-utils';
import { IManagedServiceContext } from '../_.$account.$cluster.$project.$scope.$workspace.backing-service.$service/route';
import HandleBackendResources from './handle-backend-resources';
import ManagedResources from './managed-resources';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);

  const { service } = ctx.params;
  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listManagedResource({
      ...getScopeAndProjectQuery(ctx),
      search: {
        ...getSearch(ctx),
        managedServiceName: {
          matchType: 'exact',
          exact: service,
        },
      },
    });
    if (errors) {
      throw errors[0];
    }
    return {
      resourcesData: data || {},
    };
  });

  return defer({ promise });
};

const BackingResources = () => {
  const [showBackendResourceDialog, setShowBackendResourceDialog] =
    useState<IShowDialog>(null);
  const { promise } = useLoaderData<typeof loader>();

  const { managedTemplates, backendService } =
    useOutletContext<IManagedServiceContext>();

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          resourcesData:
            fake.ConsoleListManagedResourceQuery.core_listManagedResources,
        }}
      >
        {({ resourcesData }) => {
          const resources = parseNodes(resourcesData);

          if (!resources) {
            return null;
          }
          return (
            <Wrapper
              header={{
                title: 'Resources',
                action: (
                  <Button
                    variant="primary"
                    content="Create resource"
                    prefix={<PlusFill />}
                    onClick={() => {
                      setShowBackendResourceDialog({ type: '', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                is: resources.length === 0,
                title:
                  'This is where you’ll manage your backing service resources..',
                content: (
                  <p>
                    You can create a new resource and manage the listed
                    resources.
                  </p>
                ),
                action: {
                  content: 'Create new resource',
                  prefix: <Plus />,
                  onClick: () => {
                    setShowBackendResourceDialog({ type: '', data: null });
                  },
                },
              }}
              tools={<Tools />}
            >
              <ManagedResources items={resources} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleBackendResources
        template={
          getManagedTemplate({
            templates: managedTemplates,
            kind: backendService.spec.msvcKind.kind || '',
            apiVersion: backendService.spec.msvcKind.apiVersion,
          })!
        }
        setShow={setShowBackendResourceDialog}
        show={showBackendResourceDialog}
      />
    </>
  );
};

export default BackingResources;
