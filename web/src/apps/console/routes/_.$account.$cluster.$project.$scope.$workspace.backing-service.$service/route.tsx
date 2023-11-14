import { redirect } from '@remix-run/node';
import {
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { IManagedService } from '~/console/server/gql/queries/managed-service-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { getScopeAndProjectQuery } from '~/console/server/r-utils/common';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import { IWorkspaceContext } from '../_.$account.$cluster.$project.$scope.$workspace/route';

export interface IManagedServiceContext extends IWorkspaceContext {
  backendService: IManagedService;
}

const BackendService = () => {
  const rootContext = useOutletContext<IWorkspaceContext>();
  const { backendService } = useLoaderData();
  return <Outlet context={{ ...rootContext, backendService }} />;
};

const Tabs = () => {
  const { account, cluster, project, scope, workspace, service } = useParams();
  return (
    <CommonTabs
      baseurl={`${account}/${cluster}/${project}/${scope}/${workspace}/backing-service/${service}`}
      backButton={{
        to: `/${account}/${cluster}/${project}/${scope}/${workspace}/backing-services`,
        label: 'Backend services',
      }}
      tabs={[
        {
          label: 'Monitoring',
          to: '/monitoring',
          value: '/monitoring',
        },
        {
          label: 'Logs',
          to: '/logs',
          value: '/logs',
        },
        {
          label: 'Resources',
          to: '/resources',
          value: '/resources',
        },
        {
          label: 'Maintenance',
          to: '/maintenance',
          value: '/maintenance',
        },
        {
          label: 'Settings',
          to: '/settings',
          value: '/settings',
        },
      ]}
    />
  );
};

export const handle = () => {
  return {
    navbar: <Tabs />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  const { service } = ctx.params;

  ensureClusterSet(ctx);
  ensureAccountSet(ctx);

  const api = GQLServerHandler(ctx.request).getManagedService;

  try {
    const { data, errors } = await api({
      name: service,
      ...getScopeAndProjectQuery(ctx),
    });
    if (errors) {
      logger.error(errors);
      throw errors[0];
    }
    return {
      backendService: data || {},
    };
  } catch (err) {
    return redirect(`../backing-services`);
  }
};

export default BackendService;
