import { defer } from '@remix-run/node';
import {
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { IApp } from '~/console/server/gql/queries/app-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
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
import { IWorkspaceContext } from '../_.$account.$cluster.$project.$scope.$workspace/route';

const ProjectTabs = () => {
  const { account, cluster, project, scope, workspace, app } = useParams();
  return (
    <CommonTabs
      baseurl={`/${account}/${cluster}/${project}/${scope}/${workspace}/app/${app}`}
      backButton={{
        to: `/${account}/${cluster}/${project}/${scope}/${workspace}/apps`,
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

export interface IAppContext extends IWorkspaceContext {
  app: IApp;
}

const AppOutlet = ({ app: oApp }: { app: IApp }) => {
  const rootContext = useOutletContext<IWorkspaceContext>();
  const { data: subNavData, setData: setSubNavAction } = useSubNavData();
  const { app, setApp, resetState } = useAppState();
  const [isOpen, setIsOpen] = useState(false);

  const {
    account,
    cluster,
    project,
    scope,
    workspace,
    app: appId,
  } = useParams();
  const { activePath } = useActivePath({
    parent: `/${account}/${cluster}/${project}/${scope}/${workspace}/app/${appId}/settings`,
  });

  useEffect(() => {
    resetState(oApp);
  }, [activePath]);

  const api = useConsoleApi();

  const { isLoading, submit } = useForm({
    initialValues: {},
    validationSchema: Yup.object({}),
    onSubmit: async () => {
      try {
        const { errors } = await api.updateApp({
          app: getAppIn(app),
        });
        if (errors) {
          throw errors[0];
        }
        toast.success('app updated');
        // @ts-ignore
        window.reload();
      } catch (err) {
        handleError(err);
      }
    },
  });

  // useEffect(() => {
  //   if (JSON.stringify(app) !== JSON.stringify(oApp)) {
  //     setSubNavAction({
  //       ...(subNavData || {}),
  //       show: true,
  //       content: 'View Changes',
  //       action() {
  //         setIsOpen(true);
  //       },
  //       subAction() {
  //         setApp(oApp);
  //       },
  //     });
  //   } else {
  //     setSubNavAction({
  //       ...(subNavData || {}),
  //       show: false,
  //     });
  //   }
  // }, [app, oApp]);
  return (
    <>
      <Popup.Root
        className="w-[90vw] max-w-[1440px] min-w-[1000px]"
        show={isOpen}
        onOpenChange={(v) => setIsOpen(v)}
      >
        <Popup.Header>Review Changes</Popup.Header>
        <Popup.Content>
          <DiffViewer
            oldValue={yamlDump(getAppIn(oApp)).toString()}
            newValue={yamlDump(getAppIn(app)).toString()}
            leftTitle="Previous State"
            rightTitle="New State"
            splitView
          />
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button
            loading={isLoading}
            onClick={() => {
              submit();
            }}
            content="Commit Changes"
          />
        </Popup.Footer>
      </Popup.Root>

      <Outlet context={{ ...rootContext, app: oApp }} />
    </>
  );
};

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    ensureClusterSet(ctx);
    const { app, scope, workspace, project } = ctx.params;
    try {
      const { data, errors } = await GQLServerHandler(ctx.request).getApp({
        project: {
          type: 'name',
          value: project,
        },
        name: app,
        scope: {
          value: workspace,
          type: scope === 'workspace' ? 'workspaceName' : 'environmentName',
        },
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
        return (
          <AppContextProvider initialAppState={app}>
            <SubNavDataProvider>
              <AppOutlet app={app} />
            </SubNavDataProvider>
          </AppContextProvider>
        );
      }}
    </LoadingComp>
  );
};

export default App;
