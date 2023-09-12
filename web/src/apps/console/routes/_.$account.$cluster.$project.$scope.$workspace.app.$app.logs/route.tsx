import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { listOrGrid, parseNodes } from '~/console/server/r-utils/common';
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
  const { cluster } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listProjects({
      clusterName: cluster,
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });

    if (errors) {
      throw errors[0];
    }
    return { projectsData: data };
  });

  return defer({ promise });
};

const AppLogs = () => {
  const [viewMode, setViewMode] = useState<listOrGrid>('list');

  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp data={promise}>
      {({ projectsData }) => {
        const projects = parseNodes(projectsData);
        if (!projects) {
          return null;
        }
        return (
          <Wrapper
            header={{
              title: 'Logs',
            }}
          >
            <Tools viewMode={viewMode} setViewMode={setViewMode} />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default AppLogs;
