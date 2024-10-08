import { Outlet, useOutletContext } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { cn } from '~/components/utils';
import { DiffViewer, yamlDump } from '~/console/components/diff-viewer';
import SidebarLayout from '~/console/components/sidebar-layout';
import {
  AppContextProvider,
  useAppState,
} from '~/console/page-components/app-states';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName } from '~/console/server/r-utils/common';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import { getAppIn } from '~/console/server/r-utils/resource-getter';
import { constants } from '~/console/server/utils/constants';
import { useReload } from '~/lib/client/helpers/reloader';
import useForm from '~/lib/client/hooks/use-form';
import {
  DISCARD_ACTIONS,
  UnsavedChangesProvider,
  useUnsavedChanges,
} from '~/lib/client/hooks/use-unsaved-changes';
import Yup from '~/lib/server/helpers/yup';
import { handleError } from '~/lib/utils/common';
import { registryHost } from '~/root/lib/configs/base-url.cjs';
import appFun from '../../../new-app/app-pre-submit';
import { getImageTag } from '../../../new-app/app-utils';
import { IAppContext } from '../_layout';

const navItems = [
  { label: 'General', value: 'general' },
  { label: 'Compute', value: 'compute' },
  { label: 'Environment', value: 'environment' },
  { label: 'Network', value: 'network' },
  { label: 'Advanced', value: 'advance' },
];

const Layout = () => {
  const rootContext = useOutletContext<IAppContext>();
  const {
    setHasChanges,
    setIgnorePaths,
    performAction,
    setPerformAction,
    loading,
  } = useUnsavedChanges();
  const {
    app,
    readOnlyApp,
    buildData,
    setBuildData,
    setApp,
    setReadOnlyApp,
    existingBuildId,
    setExistingBuildID,
  } = useAppState();

  const { environment, account } = useOutletContext<IAppContext>();
  const [envName, accountName, appName] = [
    parseName(environment),
    parseName(account),
    parseName(app),
  ];

  const [showDiff, setShowDiff] = useState(false);

  useEffect(() => {
    setIgnorePaths(
      navItems.map(
        (ni) =>
          `/${account}/env/${environment}/app/${appName}/settings/${ni.value}`
      )
    );
  }, []);

  const api = useConsoleApi();
  const reload = useReload();

  const { isLoading, submit } = useForm({
    initialValues: {},
    validationSchema: Yup.object({}),
    onSubmit: async () => {
      if (!environment) {
        throw new Error('Project and Environment is required!.');
      }

      const gitMode =
        app.metadata?.annotations?.[keyconstants.appImageMode] === 'git';
      // update or create build first if git image is selected
      let buildId: string | null | undefined = app.ciBuildId;
      let tagName = '';

      if (buildData && gitMode) {
        try {
          tagName = getImageTag({
            app: parseName(app),
            environment: envName,
          });
          if (!app.ciBuildId && !existingBuildId) {
            buildId = await appFun.createBuild({
              api,
              build: buildData,
            });
          }

          if (existingBuildId) {
            buildId = existingBuildId;
          }

          if (buildId) {
            await appFun.triggerBuild({ api, buildId });
          }
        } catch (err) {
          handleError(err);
          return;
        }
      }

      try {
        const { errors } = await api.updateApp({
          app: {
            ...getAppIn(app),
            ...(buildId
              ? {
                  ciBuildId: buildId,
                  spec: {
                    ...app.spec,
                    ...(gitMode
                      ? {
                          containers: [
                            {
                              ...app.spec.containers?.[0],
                              image: (() => {
                                if (existingBuildId) {
                                  return `${registryHost}/${accountName}/${
                                    buildData?.spec.registry.repo.name
                                  }:${
                                    buildData?.spec.registry.repo.tags?.[0] ||
                                    'latest'
                                  }`;
                                }
                                if (app.ciBuildId) {
                                  if (
                                    readOnlyApp.spec.containers?.[0].image.includes(
                                      constants.defaultAppRepoNameOnly
                                    )
                                  ) {
                                    return `${constants.defaultAppRepoName(
                                      accountName
                                    )}:${tagName}`;
                                  }
                                  return `${registryHost}/${accountName}/${
                                    buildData?.spec.registry.repo.name
                                  }:${
                                    buildData?.spec.registry.repo.tags?.[0] ||
                                    'latest'
                                  }`;
                                }
                                return `${constants.defaultAppRepoName(
                                  accountName
                                )}:${tagName}`;
                              })(),
                              name: 'container-0',
                            },
                          ],
                        }
                      : {}),
                  },
                }
              : {}),
          },
          envName,
        });
        if (errors) {
          throw errors[0];
        }
        toast.success('App updated successfully');
        // @ts-ignore
        setPerformAction(DISCARD_ACTIONS.INIT);
        if (!gitMode) {
          // @ts-ignore
          setBuildData(null);
        }
        setExistingBuildID(null);
        setHasChanges(false);
        reload();
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    if (loading) {
      return;
    }
    const isNotSame = JSON.stringify(app) !== JSON.stringify(rootContext.app);

    const isBuildNotSame =
      JSON.stringify(buildData) !== JSON.stringify(rootContext.app.build);

    const isBuildUndefAndNull =
      buildData === undefined && rootContext.app.build === null;

    if (isNotSame || (isBuildNotSame && !isBuildUndefAndNull)) {
      setHasChanges(true);
    } else {
      setHasChanges(false);
    }
  }, [app, rootContext.app, buildData, loading]);

  useEffect(() => {
    if (!loading) {
      setApp(rootContext.app);
      setReadOnlyApp(rootContext.app);
      // @ts-ignore
      setBuildData(rootContext.app.build);
    }
    setHasChanges(false);
  }, [rootContext.app]);

  useEffect(() => {
    if (performAction === DISCARD_ACTIONS.DISCARD_CHANGES) {
      setApp(rootContext.app);
      setReadOnlyApp(rootContext.app);
      // @ts-ignore
      setBuildData(rootContext.app.build);
      setPerformAction('');
    }
  }, [performAction]);

  return (
    <SidebarLayout navItems={navItems} parentPath="/settings">
      <Popup.Root
        className={cn('w-[90vw] max-w-[1440px]', {
          'min-w-[1000px]': showDiff,
          'min-w-[500px]': !showDiff,
        })}
        show={performAction === DISCARD_ACTIONS.VIEW_CHANGES}
        onOpenChange={(v) => setPerformAction(v)}
      >
        <Popup.Header>Commit Changes</Popup.Header>
        <Popup.Content>
          <div className="flex flex-col gap-md">
            <span className="bodyMd-medium text-text-strong">
              Please confirm if you want to update this app. This action will
              overwrite existing app details.
            </span>
            <Button
              size="sm"
              content={
                <span className="truncate text-left">
                  {showDiff
                    ? 'Hide Changes?'
                    : 'Click here to review changes before proceeding.'}
                </span>
              }
              variant="primary-plain"
              className="truncate"
              onClick={() => {
                setShowDiff(!showDiff);
              }}
            />
          </div>
          {showDiff && (
            <>
              <DiffViewer
                oldValue={yamlDump(getAppIn(rootContext.app)).toString()}
                newValue={yamlDump(getAppIn(app)).toString()}
                leftTitle="Previous State"
                rightTitle="New State"
                splitView
              />
              <DiffViewer
                oldValue={yamlDump(rootContext.app.build).toString()}
                newValue={yamlDump(buildData).toString()}
                leftTitle="Previous State"
                rightTitle="New State"
                splitView
              />
            </>
          )}
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
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

const Settings = () => {
  const rootContext = useOutletContext<IAppContext>();
  return (
    <AppContextProvider initialAppState={rootContext.app}>
      <UnsavedChangesProvider
        onProceed={({ setPerformAction }) => {
          setPerformAction?.(DISCARD_ACTIONS.DISCARD_CHANGES);
        }}
      >
        <Layout />
      </UnsavedChangesProvider>
    </AppContextProvider>
  );
};

export default Settings;
