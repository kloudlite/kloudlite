import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { project, scope, workspace } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listRouters({
      project: {
        value: project,
        type: 'name',
      },
      scope: {
        value: workspace,
        type: scope === 'workspace' ? 'workspaceName' : 'environmentName',
      },
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
  const [viewMode, setViewMode] = useState<'list' | 'grid'>('list');

  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp data={promise}>
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
                  prefix={<PlusFill />}
                  to="#TODO"
                  LinkComponent={Link}
                />
              ),
            }}
            empty={{
              is: routers.length === 0,
              title: 'This is where you’ll manage your Routers.',
              content: (
                <p>You can create a new router and manage the listed router.</p>
              ),
              action: {
                content: 'Add new router',
                prefix: <Plus />,
                LinkComponent: Link,
                to: `#TODO`,
              },
            }}
          >
            <Tools viewMode={viewMode} setViewMode={setViewMode} />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default Routers;
