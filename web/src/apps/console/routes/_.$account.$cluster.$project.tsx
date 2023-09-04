import {
  Outlet,
  useOutletContext,
  useLoaderData,
  useParams,
} from '@remix-run/react';
import { redirect } from '@remix-run/node';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { ensureAccountSet, ensureClusterSet } from '../server/utils/auth-utils';
import { CommonTabs } from '../components/common-navbar-tabs';
import { IClusterContext } from './_.$account.$cluster';
import { type IProject } from '../server/gql/queries/project-queries';

export interface IProjectContext {
  project: IProject;
}

const Project = () => {
  const rootContext = useOutletContext<IClusterContext>();
  const { project } = useLoaderData();
  return <Outlet context={{ ...rootContext, project }} />;
};

const ProjectTabs = () => {
  const { account, cluster, project } = useParams();
  return (
    <CommonTabs
      baseurl={`/${account}/${cluster}/${project}`}
      backButton={{
        to: `/${account}/${cluster}/projects`,
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
          to: '/settings/access-management',
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
