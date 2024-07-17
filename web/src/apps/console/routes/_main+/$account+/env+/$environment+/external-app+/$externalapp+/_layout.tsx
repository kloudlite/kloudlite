import { defer } from '@remix-run/node';
import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import logger from '~/lib/client/helpers/log';
import { IRemixCtx } from '~/lib/types/common';

import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { BreadcrumSlash } from '~/console/utils/commons';
import Breadcrum from '~/console/components/breadcrum';
import { IExternalApp } from '~/console/server/gql/queries/external-app-queries';
import { Truncate } from '~/root/lib/utils/common';
import { parseName } from '~/console/server/r-utils/common';
import { IEnvironmentContext } from '../../_layout';

const LocalTabs = () => {
  const { account, environment, externalapp } = useParams();
  return (
    <CommonTabs
      baseurl={`/${account}/env/${environment}/external-app/${externalapp}`}
      backButton={{
        to: `/${account}/env/${environment}/external-apps`,
        label: 'External Apps',
      }}
      tabs={[
        // {
        //   label: 'Logs & Metrics',
        //   to: '/logs-n-metrics',
        //   value: '/logs-n-metrics',
        // },
        {
          label: 'Settings',
          to: '/settings/general',
          value: '/settings',
        },
      ]}
    />
  );
};

const LocalBreadcrum = ({ data }: { data: IExternalApp }) => {
  const params = useParams();

  const { account, environment } = params;

  const { displayName } = data;
  return (
    <div className="flex flex-row items-center">
      <BreadcrumSlash />
      <span className="mx-md" />
      <Breadcrum.Button
        to={`/${account}/env/${environment}/external-apps`}
        linkComponent={Link}
        content="External Apps"
      />
      <BreadcrumSlash />
      <Breadcrum.Button
        content={<Truncate length={15}>{displayName || ''}</Truncate>}
        size="sm"
        variant="plain"
        linkComponent={Link}
        to={`/${account}/env/${environment}/external-app/${parseName(
          data
        )}/settings/general`}
      />
    </div>
  );
};

export const handle = ({ promise: { app, error } }: { promise: any }) => {
  if (error) {
    return {};
  }
  return {
    navbar: <LocalTabs />,
    breadcrum: () => <LocalBreadcrum data={app} />,
  };
};

export interface IExternalAppContext extends IEnvironmentContext {
  app: IExternalApp;
}

const ExternalAppOutlet = ({ app: oApp }: { app: IExternalApp }) => {
  const rootContext = useOutletContext<IEnvironmentContext>();

  return <Outlet context={{ ...rootContext, app: oApp }} />;
};

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { externalapp, environment } = ctx.params;
    try {
      const { data, errors } = await GQLServerHandler(
        ctx.request
      ).getExternalApp({
        envName: environment,
        name: externalapp,
      });
      if (errors) {
        throw errors[0];
      }
      return {
        app: data,
      };
    } catch (err) {
      logger.log(err);

      return {
        app: {} as IExternalApp,
        redirect: '../external-apps',
      };
    }
  });
  return defer({ promise: await promise });
};

const App = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp data={promise}>
      {({ app }) => {
        return <ExternalAppOutlet app={app} />;
      }}
    </LoadingComp>
  );
};

export default App;
