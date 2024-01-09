import { defer } from '@remix-run/node';
import {
  Outlet,
  useLoaderData,
  useLocation,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { IApp } from '~/console/server/gql/queries/app-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import logger from '~/root/lib/client/helpers/log';
import {
  SubNavDataProvider,
  useSubNavData,
} from '~/root/lib/client/hooks/use-create-subnav-action';
import { IRemixCtx } from '~/root/lib/types/common';

import { useEffect, useState } from 'react';
import { toast } from 'react-toastify';
import Popup from '~/components/molecule/popup';
import { DiffViewer, yamlDump } from '~/console/components/diff-viewer';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import {
  AppContextProvider,
  useAppState,
} from '~/console/page-components/app-states';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { getAppIn } from '~/console/server/r-utils/resource-getter';
import { useActivePath } from '~/root/lib/client/hooks/use-active-path';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { UnsavedChangesProvider } from '~/root/lib/client/hooks/use-unsaved-changes';
import { IEnvironmentContext } from '../../_layout';

const ProjectTabs = () => {
  const { account, project, environment, app } = useParams();
  return (
    <CommonTabs
      baseurl={`/${account}/${project}/${environment}/app/${app}`}
      backButton={{
        to: `/${account}/${project}/${environment}/apps`,
        label: 'Apps',
      }}
      tabs={[
        {
          label: 'Overview',
          to: '/overview',
          value: '/overview',
        },
        {
          label: 'Logs',
          to: '/logs',
          value: '/logs',
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

export interface IAppContext extends IEnvironmentContext {
  app: IApp;
}

const AppOutlet = ({ app: oApp }: { app: IApp }) => {
  const rootContext = useOutletContext<IEnvironmentContext>();

  return <Outlet context={{ ...rootContext, app: oApp }} />;
};

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { app, environment, project } = ctx.params;
    try {
      const { data, errors } = await GQLServerHandler(ctx.request).getApp({
        envName: environment,
        name: app,
        projectName: project,
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
