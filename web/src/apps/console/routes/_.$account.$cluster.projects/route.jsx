import { useState } from 'react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import Wrapper from '~/console/components/wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { useParams, useLoaderData, Link } from '@remix-run/react';
import { defer } from '@remix-run/node';
import logger from '~/root/lib/client/helpers/log';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import {
  getPagination,
  getSearch,
  parseName,
} from '~/console/server/r-urils/common';
import { parseError } from '~/root/lib/utils/common';
import ResourceList from '../../components/resource-list';
import Resources from '../_.$account.projects._index/resources';
import Tools from './tools';

const ClusterDetail = () => {
  const [viewMode, setViewMode] = useState('list');

  const { account, cluster } = useParams();
  const { promise } = useLoaderData();
  // @ts-ignore

  return (
    <LoadingComp data={promise}>
      {({ projectsData }) => {
        const projects = projectsData.edges?.map(({ node }) => node);
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

export const loader = async (ctx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { cluster } = ctx.params;

  const promise = pWrapper(async () => {
    try {
      const { data, errors } = await GQLServerHandler(ctx.request).listProjects(
        {
          clusterName: cluster,
          pagination: getPagination(ctx),
          search: getSearch(ctx),
        }
      );
      if (errors) {
        throw errors[0];
      }
      return { projectsData: data };
    } catch (err) {
      logger.error(err);
      return { error: parseError(err).message };
    }
  });

  return defer({ promise });
};

export default ClusterDetail;
