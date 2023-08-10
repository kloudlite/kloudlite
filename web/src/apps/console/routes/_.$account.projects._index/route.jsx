import { useState } from 'react';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import logger from '~/root/lib/client/helpers/log';
import Filters from '~/console/components/filters';
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
            <div className="flex flex-col">
              <Tools viewMode={viewMode} setViewMode={setViewMode} />
              <Filters
                appliedFilters={appliedFilters}
                setAppliedFilters={setAppliedFilters}
              />
            </div>
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
    }

    return {
      projectsData: data || {},
    };
  });

  return defer({ promise });
};

export default ProjectsIndex;
