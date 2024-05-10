import { defer } from '@remix-run/node';
import {
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { CommonTabs } from '~/iotconsole/components/common-navbar-tabs';
import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import { ensureAccountSet } from '~/iotconsole/server/utils/auth-utils';
import logger from '~/lib/client/helpers/log';
import { IRemixCtx } from '~/lib/types/common';

import {
  LoadingComp,
  pWrapper,
} from '~/iotconsole/components/loading-component';
import { BreadcrumSlash } from '~/iotconsole/utils/commons';
import Breadcrum from '~/iotconsole/components/breadcrum';
import { Truncate } from '~/root/lib/utils/common';
import { IApp } from '~/iotconsole/server/gql/queries/iot-app-queries';
import { IDeviceBlueprintContext } from '../../_layout';

const LocalTabs = () => {
  const { account, deviceblueprint, app } = useParams();
  return (
    <CommonTabs
      baseurl={`/${account}/deviceblueprint/${deviceblueprint}/app/${app}`}
      backButton={{
        to: `/${account}/deviceblueprint/${deviceblueprint}/apps`,
        label: 'Apps',
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

const LocalBreadcrum = ({ data }: { data: IApp }) => {
  const { displayName } = data;
  return (
    <div className="flex flex-row items-center">
      <BreadcrumSlash />
      <Breadcrum.Button
        content={<Truncate length={15}>{displayName || ''}</Truncate>}
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

export interface IAppContext extends IDeviceBlueprintContext {
  app: IApp;
}

const AppOutlet = ({ app: oApp }: { app: IApp }) => {
  const rootContext = useOutletContext<IDeviceBlueprintContext>();

  return <Outlet context={{ ...rootContext, app: oApp }} />;
};

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { app, deviceblueprint } = ctx.params;
    try {
      const { data, errors } = await GQLServerHandler(ctx.request).getIotApp({
        deviceBlueprintName: deviceblueprint,
        name: app,
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
        app: {} as IApp,
        redirect: '../apps',
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
        return <AppOutlet app={app} />;
      }}
    </LoadingComp>
  );
};

export default App;
