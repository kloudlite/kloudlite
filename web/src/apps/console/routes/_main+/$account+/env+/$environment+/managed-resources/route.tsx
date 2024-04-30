import { Plus } from '~/console/components/icons';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { IRemixCtx } from '~/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import { Button } from '~/components/atoms/button.js';
import Tools from './tools';
import ManagedResourceResourcesV2 from './managed-resources-resource-v2';

export const loader = (ctx: IRemixCtx) => {
  const { environment } = ctx.params;
  const promise = pWrapper(async () => {
    const { data: mData, errors: mErrors } = await GQLServerHandler(
      ctx.request
    ).listManagedResources({
      
      envName: environment,
    });

    if (mErrors) {
      throw mErrors[0];
    }
    return { managedResourcesData: mData };
  });
  return defer({ promise });
};

const KlOperatorServices = () => {
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
                  LinkComponent={Link}
                />
              ),
            }}
            empty={{
              is: managedResources.length === 0,
              title: 'This is where you’ll manage your Managed resources.',
              content: (
                <p>
                  You can create a new backing resource and manage the listed
                  backing resource.
                </p>
              ),
              action: {
                content: 'Create new managed resource',
                prefix: <Plus />,
                to: '../new-managed-resource',
                LinkComponent: Link,
              },
            }}
            tools={<Tools />}
          >
            <ManagedResourceResourcesV2
              items={managedResources}
              templates={[]}
            />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default KlOperatorServices;
