import { Plus } from '~/iotconsole/components/icons';
import { defer } from '@remix-run/node';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button.jsx';
import {
  LoadingComp,
  pWrapper,
} from '~/iotconsole/components/loading-component';
import Wrapper from '~/iotconsole/components/wrapper';
import { getPagination, getSearch } from '~/iotconsole/server/utils/common';
import logger from '~/root/lib/client/helpers/log';
import { ensureAccountSet } from '~/iotconsole/server/utils/auth-utils';
import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import { IRemixCtx } from '~/root/lib/types/common';
// import useActiveDevice from '~/iotconsole/hooks/use-device';
// import { useEffect } from 'react';
import Tools from './tools';
import ProjectResourcesV2 from './project-resources-v2';

export const loader = (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);

    const { data: projects, errors } = await GQLServerHandler(
      ctx.request
    ).listIotProjects({
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    console.log(projects);

    return {
      projectsData: projects || {},
    };
  });

  return defer({ promise });
};

const Projects = () => {
  // return <Wip />;
  const { account } = useParams();
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp
      data={promise}
      // skeletonData={{
      //   projectsData: fake.ConsoleListProjectsQuery.core_listProjects as any,
      // }}
    >
      {({ projectsData }) => {
        const projects = projectsData.edges?.map(({ node }) => node);

        return (
          <Wrapper
            header={{
              title: 'Projects',
              action: projects.length > 0 && (
                <Button
                  variant="primary"
                  content="Create Project"
                  prefix={<Plus />}
                  to={`/${account}/new-project`}
                  linkComponent={Link}
                />
              ),
            }}
            empty={{
              is: projects.length === 0,
              title: 'you have not added any project yet.',
              content: (
                <p>
                  please add some cloud providers to start creating cluster.
                </p>
              ),
              action: {
                content: 'Add Project',
                prefix: <Plus />,
                linkComponent: Link,
                to: `/${account}/new-project`,
              },
            }}
            tools={<Tools />}
            pagination={{
              pageInfo: projectsData.pageInfo,
            }}
          >
            <ProjectResourcesV2 items={projects} />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default Projects;
