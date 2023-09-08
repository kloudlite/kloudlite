import { useState } from 'react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import Wrapper from '~/console/components/wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { useParams, useLoaderData, Link } from '@remix-run/react';
import { defer } from '@remix-run/node';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import {
  listOrGrid,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import ResourceList from '../../components/resource-list';
import Resources from '../_.$account.projects._index/resources';
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

const ClusterDetail = () => {
  const [viewMode, setViewMode] = useState<listOrGrid>('list');

  const { account, cluster } = useParams();
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
              title: 'Projects',
              action: projects.length > 0 && (
                <Button
                  variant="primary"
                  content="Create Project"
                  prefix={<PlusFill />}
                  to={`/onboarding/${account}/${cluster}/new-project`}
                  LinkComponent={Link}
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
                prefix: <Plus />,
                LinkComponent: Link,
                to: `/${account}/new-project`,
              },
            }}
          >
            <Tools viewMode={viewMode} setViewMode={setViewMode} />
            <ResourceList mode={viewMode} linkComponent={Link} prefetchLink>
              {projects.map((project) => {
                return (
                  <ResourceList.ResourceItem
                    to={`/${account}/${project.clusterName}/${parseName(
                      project
                    )}/workspaces`}
                    key={parseName(project)}
                    textValue={parseName(project)}
                  >
                    <Resources item={project} />
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

export default ClusterDetail;
