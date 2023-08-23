import { useState } from 'react';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import logger from '~/root/lib/client/helpers/log';
import Wrapper from '~/console/components/wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import {
  getPagination,
  getSearch,
  parseName,
  parseNodes,
} from '~/console/server/r-urils/common';
import { defer } from 'react-router-dom';
import HandleScope, { SCOPE } from '~/console/page-components/new-scope';
import ResourceList from '../../components/resource-list';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '../../server/utils/auth-utils';
import Tools from './tools';
import Resources from './resources';

const Workspaces = () => {
  const [viewMode, setViewMode] = useState('list');
  const [showAddWS, setShowAddWS] = useState(null);

  const { account, project, cluster } = useParams();
  const { promise } = useLoaderData();
  return (
    <>
      <LoadingComp data={promise}>
        {({ workspacesData }) => {
          const workspaces = parseNodes(workspacesData);

          if (!workspaces) {
            return null;
          }

          return (
            <Wrapper
              header={{
                title: 'Workspaces',
                action: (
                  <Button
                    variant="primary"
                    content="Create Workspace"
                    prefix={PlusFill}
                    onClick={() => {
                      setShowAddWS({ type: 'add', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                is: workspaces.length === 0,
                title: 'This is where you’ll manage your workspaces.',
                content: (
                  <p>
                    You can create a new workspace and manage the listed
                    workspaces.
                  </p>
                ),
                action: {
                  content: 'Create new workspace',
                  prefix: Plus,
                  onClick: () => {
                    setShowAddWS({ type: 'add', data: null });
                  },
                },
              }}
            >
              <Tools viewMode={viewMode} setViewMode={setViewMode} />
              <ResourceList mode={viewMode} linkComponent={Link} prefetchLink>
                {workspaces.map((ws) => (
                  <ResourceList.ResourceItem
                    to={`/${account}/${cluster}/${project}/workspace/${parseName(
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
        scope={SCOPE.WORKSPACE}
      />
    </>
  );
};

export const loader = async (ctx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { project } = ctx.params;
  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listWorkspaces(
      {
        project: {
          type: 'name',
          value: project,
        },
        pagination: getPagination(ctx),
        search: getSearch(ctx),
      }
    );
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      workspacesData: data || {},
    };
  });

  return defer({ promise });
};

export default Workspaces;
