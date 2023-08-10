import { useState } from 'react';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import logger from '~/root/lib/client/helpers/log';
import Filters from '~/console/components/filters';
import Wrapper from '~/console/components/wrapper';
import ResourceList from '../../components/resource-list';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import { dummyData } from '../../dummy/data';
import { ensureAccountSet } from '../../server/utils/auth-utils';
import Tools from './tools';
import Resources from './resources';

const ProjectsIndex = () => {
  const [appliedFilters, setAppliedFilters] = useState(
    dummyData.appliedFilters
  );

  const [viewMode, setViewMode] = useState('list');

  const { account } = useParams();
  const { projectsData, _clustersCount } = useLoaderData();
  const [projects, _setProjects] = useState(
    projectsData.edges?.map(({ node }) => node)
  );

  return (
    <Wrapper
      header={{
        title: 'Projects',
        action: projects.length > 0 && (
          <Button
            variant="primary"
            content="Create Project"
            prefix={PlusFill}
            href={`/${account}/new-project`}
            LinkComponent={Link}
          />
        ),
      }}
      empty={{
        is: projects.length === 0,
        title: 'This is where you’ll manage your projects.',
        content: (
          <p>You can create a new project and manage the listed project.</p>
        ),
        action: {
          content: 'Add new projects',
          prefix: Plus,
          LinkComponent: Link,
          href: `/${account}/new-project`,
        },
      }}
    >
      <div className="flex flex-col">
        <Tools viewMode={viewMode} setViewMode={setViewMode} />
        <Filters
          appliedFilters={appliedFilters}
          setAppliedFilters={setAppliedFilters}
        />
      </div>
      <ResourceList mode={viewMode}>
        {projects.map((project) => (
          <ResourceList.ResourceItem key={project.id} textValue={project.id}>
            <Resources {...project} />
          </ResourceList.ResourceItem>
        ))}
      </ResourceList>
    </Wrapper>
  );
};

export const restActions = async (ctx) => {
  const { data: clusters, errors: cErrors } = await GQLServerHandler(
    ctx.request
  ).clustersCount({});

  if (cErrors) {
    logger.error(cErrors[0]);
  }

  const { data, errors } = await GQLServerHandler(ctx.request).listProjects({});
  if (errors) {
    logger.error(errors[0]);
  }

  return {
    projectsData: data || {},
    clustersCount: clusters?.totalCount || 0,
  };
};
export const loader = async (ctx) => {
  return ensureAccountSet(ctx) || restActions(ctx);
};

export default ProjectsIndex;
