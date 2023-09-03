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
import { ensureAccountSet } from '../../server/utils/auth-utils';
import Tools from './tools';
import Resources from './resources';

const ProjectsIndex = () => {
  const [viewMode, setViewMode] = useState('list');

  const { account } = useParams();
  const { promise } = useLoaderData();
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
                  to={`/${account}/new-project`}
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
  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listProjects({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    const { data: clusters, errors: e } = await GQLServerHandler(
      ctx.request
    ).listProjects({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (e) {
      logger.error(e[0]);
      throw e[0];
    }

    // if projects not found check cluster and found then reutur
    if (data.totalCount || clusters?.totalCount) {
      return {
        projectsData: data || {},
        clustersData: clusters || {},
        cloudProviderCount: -1,
      };
    }

    const { data: cp, errors: e2 } = await GQLServerHandler(
      ctx.request
    ).listProjects({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (e2) {
      logger.error(e2[0]);
      throw e2[0];
    }

    // if projects and clusters not present return cloudprovider count
    return {
      projectsData: data || {},
      clustersData: clusters || {},
      cloudProviderCount: cp?.totalCount || 0,
    };
  });

  return defer({ promise });
};

export default ProjectsIndex;
