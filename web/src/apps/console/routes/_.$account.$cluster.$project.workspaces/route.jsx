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
} from '~/console/server/r-urils/common';
import { defer } from 'react-router-dom';
import ResourceList from '../../components/resource-list';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '../../server/utils/auth-utils';
import Tools from './tools';
import Resources from './resources';
import HandleProvider from './handle-provider';

const Workspaces = () => {
  const [viewMode, setViewMode] = useState('list');
  const [showAddProvider, setShowAddProvider] = useState(null);

  const { account } = useParams();
  const { promise } = useLoaderData();
  return (
    <>
      <LoadingComp data={promise}>
        {({ workspacesData }) => {
          const projects = workspacesData.edges?.map(({ node }) => node);
          if (!projects) {
            return null;
          }
          return (
            <Wrapper
              header={{
                title: 'Workspaces',
                action: projects.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create Workspace"
                    prefix={PlusFill}
                    onClick={() => {
                      setShowAddProvider({ type: 'add', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                is: projects.length === 0,
                title: 'This is where you’ll manage your projects.',
                content: (
                  <p>
                    You can create a new project and manage the listed project.
                  </p>
                ),
                action: {
                  content: 'Add new projects',
                  prefix: Plus,
                  LinkComponent: Link,
                  href: `/${account}/new-project`,
                },
              }}
            >
              <Tools viewMode={viewMode} setViewMode={setViewMode} />
              <ResourceList mode={viewMode} linkComponent={Link} prefetchLink>
                {projects.map((project) => (
                  <ResourceList.ResourceItem
                    to={`/${account}/projects/${parseName(project)}`}
                    key={parseName(project)}
                    textValue={parseName(project)}
                  >
                    <Resources item={project} />
                  </ResourceList.ResourceItem>
                ))}
              </ResourceList>
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleProvider show={showAddProvider} setShow={setShowAddProvider} />
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
        namespace: project,
        pagination: getPagination(ctx),
        search: getSearch(ctx),
      }
    );
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    // if projects and clusters not present return cloudprovider count
    return {
      workspacesData: data || {},
    };
  });

  return defer({ promise });
};

export default Workspaces;
