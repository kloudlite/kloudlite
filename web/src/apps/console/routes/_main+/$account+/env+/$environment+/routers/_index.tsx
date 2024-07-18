import { Plus } from '~/console/components/icons';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
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
import { useState } from 'react';
import fake from '~/root/fake-data-generator/fake';
import { Button } from '~/components/atoms/button';
import Tools from './tools';
import HandleRouter from './handle-router';
import RouterResourcesV2 from './router-resources-V2';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { environment } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listRouters({
      envName: environment,
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      throw errors[0];
    }
    return { routersData: data };
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
          routersData: fake.ConsoleListRoutersQuery.core_listRouters as any,
        }}
      >
        {({ routersData }) => {
          const routers = parseNodes(routersData);
          if (!routers) {
            return null;
          }
          return (
            <Wrapper
              header={{
                title: 'Routers',
                action: routers.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create Router"
                    prefix={<Plus />}
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                is: routers.length === 0,
                title: 'This is where you’ll manage your Routers.',
                content: (
                  <p>
                    You can create a new router and manage the listed router.
                  </p>
                ),
                action: {
                  content: 'Add new router',
                  prefix: <Plus />,
                  onClick: () => {
                    setVisible(true);
                  },
                },
              }}
              tools={<Tools />}
            >
              <RouterResourcesV2 items={routers} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleRouter {...{ visible, setVisible, isUpdate: false }} />
    </>
  );
};

export default Routers;
