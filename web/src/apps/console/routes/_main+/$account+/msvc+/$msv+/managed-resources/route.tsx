import { defer } from '@remix-run/node';
import { Link, useLoaderData, useOutletContext } from '@remix-run/react';
import { Button } from '@kloudlite/design-system/atoms/button';
import { EmptyManagedResourceImage } from '~/console/components/empty-resource-images';
import { Plus } from '~/console/components/icons';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { IAccountContext } from '~/console/routes/_main+/$account+/_layout';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import ManagedResourceResourcesV2 from './managed-resources-resource-v2';
import Tools from './tools';

export const loader = (ctx: IRemixCtx) => {
  const { msv } = ctx.params;
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);

    const { data: mData, errors: mErrors } = await GQLServerHandler(
      ctx.request
    ).listManagedResources({
      search: {
        ...getSearch(ctx),
        managedServiceName: { matchType: 'exact', exact: msv },
      },
      pq: getPagination(ctx),
    });

    if (mErrors) {
      throw mErrors[0];
    }
    return { managedResourcesData: mData };
  });
  return defer({ promise });
};

const KlOperatorServices = () => {
  const { msvtemplates } = useOutletContext<IAccountContext>();
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp
      data={promise}
      skeletonData={{
        managedResourcesData: fake.ConsoleListManagedResourcesQuery
          .core_listManagedResources as any,
      }}
    >
      {({ managedResourcesData }) => {
        const managedResources = parseNodes(managedResourcesData);

        return (
          <Wrapper
            header={{
              title: 'Managed resources',
              action: managedResources.length > 0 && (
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
              is: managedResources.length === 0,
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
            pagination={managedResourcesData}
          >
            <ManagedResourceResourcesV2
              items={managedResources}
              templates={msvtemplates}
            />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default KlOperatorServices;
