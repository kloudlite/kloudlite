import { defer } from '@remix-run/node';
import { Link, useLoaderData, useOutletContext } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
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
import HandleManagedResourceV2 from './handle-managed-resource-v2';
import ManagedResourceResourcesV2 from './managed-resources-resource-v2';
import Tools from './tools';

export const loader = (ctx: IRemixCtx) => {
  const { environment } = ctx.params;
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);

    const { data: mData, errors: mErrors } = await GQLServerHandler(
      ctx.request
    ).listImportedManagedResources({
      envName: environment,
      search: getSearch(ctx),
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
  const [visible, setVisible] = useState(false);

  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
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
                title: 'Imported Managed Resources',
                action: managedResources.length > 0 && (
                  <Button
                    variant="primary"
                    content="Import Managed Resource"
                    prefix={<Plus />}
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                image: <EmptyManagedResourceImage />,
                is: managedResources.length === 0,
                title: 'This is where you’ll manage your Managed resources.',
                content: (
                  <p>
                    You can import a new managed resource and manage the listed
                    managed resource.
                  </p>
                ),
                action: {
                  content: 'Import Managed Resource',
                  prefix: <Plus />,
                  onClick: () => {
                    setVisible(true);
                  },
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
      <HandleManagedResourceV2
        {...{
          visible,
          setVisible,
          isUpdate: false,
        }}
      />
    </>
  );
};

export default KlOperatorServices;
