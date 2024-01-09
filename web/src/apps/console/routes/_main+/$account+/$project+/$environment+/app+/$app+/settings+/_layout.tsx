import { Outlet, useOutletContext, useParams } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import {
  UnsavedChangesProvider,
  useUnsavedChanges,
} from '~/root/lib/client/hooks/use-unsaved-changes';
import Popup from '~/components/molecule/popup';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { getAppIn } from '~/console/server/r-utils/resource-getter';
import useForm from '~/root/lib/client/hooks/use-form';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useEffect } from 'react';
import {
  AppContextProvider,
  useAppState,
} from '~/console/page-components/app-states';
import Yup from '~/root/lib/server/helpers/yup';
import { DiffViewer, yamlDump } from '~/console/components/diff-viewer';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import { IAppContext } from '../route';

const navItems = [
  { label: 'General', value: 'general' },
  { label: 'Compute', value: 'compute' },
  { label: 'Environment', value: 'environment' },
  { label: 'Network', value: 'network' },
  { label: 'Advance', value: 'advance' },
];

const Layout = () => {
  const rootContext = useOutletContext<IAppContext>();
  const {
    setHasChanges,
    setIgnorePaths,
    performAction,
    setPerformAction,
    setLoading,
  } = useUnsavedChanges();
  const { app, setApp } = useAppState();

  const { account, project, environment, app: appId } = useParams();

  useEffect(() => {
    setIgnorePaths(
      navItems.map(
        (ni) =>
          `/${account}/${project}/${environment}/app/${appId}/settings/${ni.value}`
      )
    );
  }, []);

  const api = useConsoleApi();
  const reload = useReload();

  const { isLoading, submit } = useForm({
    initialValues: {},
    validationSchema: Yup.object({}),
    onSubmit: async () => {
      if (!project || !environment) {
        throw new Error('Project and Environment is required!.');
      }
      try {
        const { errors } = await api.updateApp({
          app: getAppIn(app),
          envName: environment,
          projectName: project,
        });
        if (errors) {
          throw errors[0];
        }
        toast.success('App updated successfully');
        // @ts-ignore
        setPerformAction('');
        reload();
        setHasChanges(false);
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    const isNotSame = JSON.stringify(app) !== JSON.stringify(rootContext.app);

    if (isNotSame) {
      setHasChanges(true);
    } else {
      setHasChanges(false);
    }
  }, [app, rootContext.app]);

  useEffect(() => {
    setApp(rootContext.app);
    setLoading(false);
  }, [rootContext.app]);

  useEffect(() => {
    if (performAction === 'discard-changes') {
      setApp(rootContext.app);
      setPerformAction('');
    }
  }, [performAction]);

  return (
    <SidebarLayout navItems={navItems} parentPath="/settings">
      <Popup.Root
        className="w-[90vw] max-w-[1440px] min-w-[1000px]"
        show={performAction === 'view-changes'}
        onOpenChange={(v) => setPerformAction(v)}
      >
        <Popup.Header>Review Changes</Popup.Header>
        <Popup.Content>
          <DiffViewer
            oldValue={yamlDump(getAppIn(rootContext.app)).toString()}
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
              setLoading(true);
            }}
            content="Commit Changes"
          />
        </Popup.Footer>
      </Popup.Root>
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

const Settings = () => {
  const rootContext = useOutletContext<IAppContext>();
  if (!rootContext.app.metadata?.annotations?.[keyconstants.description]) {
    rootContext.app = {
      ...rootContext.app,
      // @ts-ignore
      metadata: {
        ...(rootContext.app.metadata || {}),
        annotations: {
          ...(rootContext.app.metadata?.annotations || {}),
          [keyconstants.description]: '',
        },
      },
    };
  }

  return (
    <AppContextProvider initialAppState={rootContext.app}>
      <UnsavedChangesProvider>
        <Layout />
      </UnsavedChangesProvider>
    </AppContextProvider>
  );
};

export default Settings;
