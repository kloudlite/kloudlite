import { redirect } from '@remix-run/node';
import {
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import logger from '~/root/lib/client/helpers/log';
import { SubNavDataProvider } from '~/root/lib/client/hooks/use-create-subnav-action';
import { IRemixCtx } from '~/root/lib/types/common';
import { IClusterContext } from '../../infra+/$cluster+/_layout';
import { IProject } from '~/console/server/gql/queries/project-queries';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { ensureAccountSet, ensureClusterSet } from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';

export interface IProjectContext extends IClusterContext {
  project: IProject;
}

const Project = () => {
  const rootContext = useOutletContext<IClusterContext>();
  const { project } = useLoaderData();
  return (
    <SubNavDataProvider>
      <Outlet context={{ ...rootContext, project }} />
    </SubNavDataProvider>
  );
};

const ProjectTabs = () => {
  const { account, cluster, project } = useParams();
  return (
    <CommonTabs
      baseurl={`/${account}/${cluster}/${project}`}
      backButton={{
        to: `/${account}/projects`,
        label: 'Projects',
      }}
      tabs={[
        {
          label: 'Environments',
          to: '/environments',
          value: '/environments',
        },
        {
          label: 'Workspaces',
          to: '/workspaces',
          value: '/workspaces',
        },

        {
          label: 'Settings',
          to: '/settings/general',
          value: '/settings',
        },
      ]}
    />
  );
};

export const handle = () => {
  return {
    navbar: <ProjectTabs />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { account, project, cluster } = ctx.params;
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getProject({
      name: project,
    });
    if (errors) {
      throw errors[0];
    }
    return {
      project: data || {},
    };
  } catch (err) {
    logger.log(err);
    return redirect(`/${account}/${cluster}/projects`);
  }
};

export default Project;
