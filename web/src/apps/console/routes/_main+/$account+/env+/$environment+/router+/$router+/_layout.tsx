import {
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { IRouter } from '~/console/server/gql/queries/router-queries';
import { redirect } from '@remix-run/node';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import logger from '~/lib/client/helpers/log';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { IRemixCtx } from '~/lib/types/common';
import { IEnvironmentContext } from '../../_layout';

export interface IRouterContext extends IEnvironmentContext {
  router: IRouter;
}
const tabs = [
  {
    label: <span className="flex flex-row items-center gap-lg">Routes</span>,
    to: '/routes',
    value: '/routes',
  },
  {
    label: <span className="flex flex-row items-center gap-lg">Settings</span>,
    to: '/settings',
    value: '/settings',
  },
];

const Routes = () => {
  const rootContext = useOutletContext<IEnvironmentContext>();
  const { router } = useLoaderData();
  return (
    <div>
      <Outlet context={{ ...rootContext, router }} />
    </div>
  );
};

const Tabs = () => {
  const { account, environment, router } = useParams();

  return (
    <CommonTabs
      baseurl={`/${account}/env/${environment}/router/${router}`}
      backButton={{
        label: 'Routers',
        to: `/${account}/env/${environment}/routers`,
      }}
      tabs={tabs}
    />
  );
};

export const handle = () => {
  return {
    navbar: <Tabs />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  const { environment,  router, account } = ctx.params;
  ensureAccountSet(ctx);

  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getRouter({
      envName: environment,
      
      name: router,
    });

    if (errors) {
      logger.error(errors);
      throw errors[0];
    }

    return {
      router: data || {},
    };
  } catch (err) {
    return redirect(`/${account}/env/${environment}/routers/`);
  }
};

export default Routes;
