import { useState } from 'react';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import Wrapper from '~/console/components/wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { parseName, parseNodes } from '~/console/server/r-urils/common';
import { defer } from '@remix-run/node';
import HandleScope, { SCOPE } from '~/console/page-components/new-scope';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { IWorkspace } from '~/console/server/gql/queries/workspace-queries';
import ResourceList from '../../components/resource-list';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '../../server/utils/auth-utils';
import Tools from './tools';
import Resources from './resources';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { project } = ctx.params;
  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listEnvironments({
      project: {
        type: 'name',
        value: project,
      },
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });

    if (errors) {
      throw errors[0];
    }

    return {
      environmentData: data || {},
    };
  });

  return defer({ promise });
};

const Workspaces = () => {
  const [viewMode, setViewMode] = useState<'list' | 'grid'>('list');
  const [showAddWS, setShowAddWS] = useState<{
    type: 'add' | 'edit';
    data: null | IWorkspace;
  } | null>(null);

  const { account, project, cluster } = useParams();
  const { promise } = useLoaderData<typeof loader>();
  return (
    <>
      <LoadingComp data={promise}>
        {({ environmentData }) => {
          const environments = parseNodes(environmentData);

          if (!environments) {
            return null;
          }

          return (
            <Wrapper
              header={{
                title: 'Environments',
                action: (
                  <Button
                    variant="primary"
                    content="Create Environment"
                    prefix={<PlusFill />}
                    onClick={() => {
                      setShowAddWS({ type: 'add', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                is: environments.length === 0,
                title: 'This is where you’ll manage your environment.',
                content: (
                  <p>
                    You can create a new workspace and manage the listed
                    workspaces.
                  </p>
                ),
                action: {
                  content: 'Create new environment',
                  prefix: <Plus />,
                  onClick: () => {
                    setShowAddWS({ type: 'add', data: null });
                  },
                },
              }}
            >
              <Tools viewMode={viewMode} setViewMode={setViewMode} />
              <ResourceList mode={viewMode} linkComponent={Link} prefetchLink>
                {environments.map((ws) => (
                  <ResourceList.ResourceItem
                    to={`/${account}/${cluster}/${project}/environment/${parseName(
                      ws
                    )}`}
                    key={parseName(ws)}
                    textValue={parseName(ws)}
                  >
                    <Resources item={ws} />
                  </ResourceList.ResourceItem>
                ))}
              </ResourceList>
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleScope
        show={showAddWS}
        setShow={setShowAddWS}
        scope={SCOPE.ENVIRONMENT}
      />
    </>
  );
};
export default Workspaces;
