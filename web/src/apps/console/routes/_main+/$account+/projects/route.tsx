import { Plus } from '~/console/components/icons';
import { defer } from '@remix-run/node';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { getPagination, getSearch } from '~/console/server/utils/common';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
// import useActiveDevice from '~/console/hooks/use-device';
// import { useEffect } from 'react';
import Tools from './tools';
import ProjectResourcesV2 from './project-resources-v2';

export const loader = (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);

    const { data: projects, errors } = await GQLServerHandler(
      ctx.request
    ).listProjects({
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    if (projects.edges.length === 0) {
      const { data: clusters, errors } = await GQLServerHandler(
        ctx.request
      ).listClusters({});
      if (errors) {
        logger.error(errors[0]);
        throw errors[0];
      }

      if (clusters.edges.length === 0) {
        const { data: secrets, errors } = await GQLServerHandler(
          ctx.request
        ).listProviderSecrets({});
        if (errors) {
          logger.error(errors[0]);
          throw errors[0];
        }

        return {
          projectsData: projects || {},
          clustersCount: 0,
          cloudProviderSecretsCount: secrets.edges.length,
        };
      }

      return {
        projectsData: projects || {},
        clustersCount: clusters.edges.length,
        cloudProviderSecretsCount: -1,
      };
    }

    return {
      projectsData: projects || {},
      clustersCount: -1,
      cloudProviderSecretsCount: -1,
    };
  });

  return defer({ promise });
};

const Projects = () => {
  // return <Wip />;
  const { account } = useParams();
  const { promise } = useLoaderData<typeof loader>();

  const getEmptyState = ({
    projectLength,
    clustersLength,
    secretsLength,
  }: {
    projectLength: number;
    clustersLength: number;
    secretsLength: number;
  }) => {
    if (secretsLength === 0) {
      return {
        is: true,
        title: 'please setup your cloud provider and cluster first',
        content: (
          <p>
            you need to setup your cluster and cloud provider first, before
            starting working with projects
          </p>
        ),
        action: {
          content: 'Setup Cloud Provider and Cluster',
          prefix: <Plus />,
          LinkComponent: Link,
          to: `/onboarding/${account}/new-cloud-provider`,
        },
      };
    }

    if (clustersLength === 0) {
      return {
        is: true,
        title: 'Setup your cluster first',
        content: (
          <p>
            you need to setup your cluster first, before starting working with
            projects
          </p>
        ),
        action: {
          content: 'Setup Cluster',
          prefix: <Plus />,
          LinkComponent: Link,
          to: `/${account}/new-cluster`,
        },
      };
    }

    if (projectLength === 0) {
      return {
        is: true,
        title: 'Create your first project',
        content: (
          <p>You can create a new project and manage the listed project.</p>
        ),
        action: {
          content: 'create projects',
          prefix: <Plus />,
          LinkComponent: Link,
          to: `/${account}/new-project`,
        },
      };
    }

    return {
      is: false,
      title: 'This is where you’ll manage your projects.',
      content: (
        <p>You can create a new project and manage the listed project.</p>
      ),
      action: {
        content: 'Add new projects',
        prefix: <Plus />,
        LinkComponent: Link,
        to: `/${account}/new-project`,
      },
    };
  };

  // const p = useActiveDevice();
  //
  // useEffect(() => {
  //   console.log(p);
  // }, [p]);

  return (
    <LoadingComp
      data={promise}
      skeletonData={{
        projectsData: fake.ConsoleListProjectsQuery.core_listProjects as any,
        clustersCount: 1,
        cloudProviderSecretsCount: 1,
      }}
    >
      {({ projectsData, clustersCount, cloudProviderSecretsCount }) => {
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
                  prefix={<Plus />}
                  to={`/${account}/new-project`}
                  LinkComponent={Link}
                />
              ),
            }}
            empty={getEmptyState({
              projectLength: projects.length,
              clustersLength: clustersCount,
              secretsLength: cloudProviderSecretsCount,
            })}
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
