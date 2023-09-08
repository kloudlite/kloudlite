import { useState } from 'react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import Wrapper from '~/console/components/wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { useLoaderData, Link } from '@remix-run/react';
import { defer } from '@remix-run/node';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { parseName, parseNodes } from '~/console/server/r-utils/common';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import ResourceList from '../../components/resource-list';
import Resources from '../_.$account.projects._index/resources';
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
            <ResourceList mode={viewMode} linkComponent={Link} prefetchLink>
              {routers.map((router) => {
                return (
                  <ResourceList.ResourceItem
                    to="#TODO"
                    key={parseName(router)}
                    textValue={parseName(router)}
                  >
                    <Resources item={router} />
                  </ResourceList.ResourceItem>
                );
              })}
            </ResourceList>
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default Routers;
