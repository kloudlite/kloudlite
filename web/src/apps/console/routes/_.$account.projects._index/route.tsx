import { useState } from 'react';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import logger from '~/root/lib/client/helpers/log';
import Wrapper from '~/console/components/wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { defer } from '@remix-run/node';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { parseName } from '~/console/server/r-utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import ResourceList from '../../components/resource-list';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import { ensureAccountSet } from '../../server/utils/auth-utils';
import Tools from './tools';
import Resources from './resources';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data: projects, errors } = await GQLServerHandler(
      ctx.request
    ).listProjects({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      projectsData: projects || {},
    };
  });

  return defer({ promise });
};

const Projects = () => {
  const [viewMode, setViewMode] = useState('list');

  const { account } = useParams();
  const { promise } = useLoaderData<typeof loader>();
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

export default Projects;
