import { Plus, PlusFill } from '@jengaicons/react';
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
// import useActiveDevice from '~/console/hooks/use-device';
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

  // const p = useActiveDevice();
  //
  // useEffect(() => {
  //   console.log(p);
  // }, [p]);

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
                  prefix={<PlusFill />}
                  to={`/${account}/new-project`}
                  LinkComponent={Link}
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
                LinkComponent: Link,
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
