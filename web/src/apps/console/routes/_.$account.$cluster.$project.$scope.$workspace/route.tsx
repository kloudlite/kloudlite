import { ChevronDown, Plus, Search } from '@jengaicons/react';
import { redirect } from '@remix-run/node';
import {
  Outlet,
  useLoaderData,
  useNavigate,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { useEffect, useState } from 'react';
import Skeleton from 'react-loading-skeleton';
import OptionList from '~/components/atoms/option-list';
import Breadcrum from '~/console/components/breadcrum';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import {
  BlackProdLogo,
  BlackWorkspaceLogo,
} from '~/console/components/commons';
import HandleScope, { SCOPE } from '~/console/page-components/new-scope';
import { IManagedServiceTemplates } from '~/console/server/gql/queries/managed-service-queries';
import { type IWorkspace } from '~/console/server/gql/queries/workspace-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  getScopeAndProjectQuery,
  parseName,
  parseNodes,
  wsOrEnv,
} from '~/console/server/r-utils/common';
import {
  ensureAccountClientSide,
  ensureAccountSet,
  ensureClusterClientSide,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import logger from '~/root/lib/client/helpers/log';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { SubNavDataProvider } from '~/root/lib/client/hooks/use-create-subnav-action';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { IRemixCtx } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import { IProjectContext } from '../_.$account.$cluster.$project';

export interface IWorkspaceContext extends IProjectContext {
  workspace: IWorkspace;
  managedTemplates: IManagedServiceTemplates;
}

const Workspace = () => {
  const rootContext = useOutletContext<IProjectContext>();
  const { workspace, managedTemplates } = useLoaderData();

  return (
    <SubNavDataProvider>
      <Outlet context={{ ...rootContext, workspace, managedTemplates }} />
    </SubNavDataProvider>
  );
};

const WorkspaceTabs = () => {
  const { account, scope, cluster, project, workspace } = useParams();
  return (
    <CommonTabs
      backButton={{
        to: `/${account}/${cluster}/${project}/${
          scope === 'workspace' ? 'workspaces' : 'environments'
        }`,
        label: scope === 'workspace' ? 'Workspaces' : 'Environments',
      }}
      baseurl={`/${account}/${cluster}/${project}/${scope}/${workspace}`}
      tabs={[
        {
          label: 'Apps',
          to: '/apps',
          value: '/apps',
        },
        {
          label: 'Routers',
          to: '/routers',
          value: '/routers',
        },
        {
          label: 'Config & Secrets',
          to: '/cs/configs',
          value: '/cs',
        },
        {
          label: 'Jobs & Crons',
          to: '/jc/task',
          value: '/jc',
        },
        {
          label: 'Backing services',
          to: '/backing-services',
          value: '/backing-services',
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
// @ts-ignore
const CurrentBreadcrum = ({ workspace }: { workspace: IWorkspace }) => {
  const params = useParams();
  const { account, cluster, project, scope } = params;

  const [showPopup, setShowPopup] = useState<any>(null);
  const [activeTab, setActiveTab] = useState<wsOrEnv>(
    (scope as wsOrEnv) || 'environment'
  );
  const [workspaces, setWorkspaces] = useState<IWorkspace[]>([]);
  const [environments, setEnvironments] = useState<IWorkspace[]>([]);

  const api = useAPIClient();
  const [search, setSearch] = useState('');

  const [isLoading, setIsLoading] = useState(false);

  useDebounce(
    async () => {
      ensureClusterClientSide(params);
      ensureAccountClientSide(params);
      const listApi =
        activeTab === SCOPE.ENVIRONMENT
          ? api.listEnvironments
          : api.listWorkspaces;
      try {
        setIsLoading(true);
        const { data, errors } = await listApi({
          project: getScopeAndProjectQuery({ params }).project,
        });
        if (errors) {
          throw errors[0];
        }

        if (activeTab === SCOPE.ENVIRONMENT) {
          setEnvironments(parseNodes(data));
        } else {
          setWorkspaces(parseNodes(data));
        }
      } catch (err) {
        handleError(err);
      } finally {
        setIsLoading(false);
      }
    },
    300,
    [search, activeTab]
  );

  useEffect(() => {
    setIsLoading(true);
  }, [activeTab]);

  const navigate = useNavigate();

  return (
    <>
      <OptionList.Root>
        <OptionList.Trigger>
          <Breadcrum.Button
            content={workspace.displayName}
            suffix={<ChevronDown />}
          />
        </OptionList.Trigger>
        <OptionList.Content className="!pt-0 !pb-md" align="center">
          <div className="p-[3px] pb-0">
            <OptionList.TextInput
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              prefixIcon={<Search />}
              placeholder="Search"
              compact
              className="border-0 rounded-none"
            />
          </div>
          <OptionList.Separator />
          <OptionList.Tabs.Root
            value={activeTab}
            size="sm"
            className="!overflow-x-visible"
            onChange={(v) => setActiveTab(v as wsOrEnv)}
          >
            <OptionList.Tabs.Tab
              prefix={<BlackWorkspaceLogo />}
              label="Workspaces"
              value="workspace"
            />
            <OptionList.Tabs.Tab
              prefix={<BlackProdLogo />}
              label="Environments"
              value="environment"
            />
          </OptionList.Tabs.Root>

          {isLoading ? (
            <Skeleton
              count={Math.max(
                [...(activeTab === 'workspace' ? environments : workspaces)]
                  .length,
                1
              )}
              height={25}
            />
          ) : (
            [...(activeTab === 'environment' ? environments : workspaces)].map(
              (item) => {
                return (
                  <OptionList.Item
                    onClick={() => {
                      navigate(
                        `/${account}/${cluster}/${project}/${
                          activeTab === 'environment'
                            ? 'environment'
                            : 'workspace'
                        }/${parseName(item)}/apps`
                      );
                    }}
                    key={parseName(item)}
                  >
                    {item.displayName}
                  </OptionList.Item>
                );
              }
            )
          )}

          <OptionList.Separator />
          <OptionList.Item
            className="text-text-primary"
            onClick={() => setShowPopup({ type: 'add' })}
          >
            <Plus size={16} />{' '}
            <span>
              {activeTab === 'workspace' ? 'New Workspace' : 'New Environment'}
            </span>
          </OptionList.Item>
        </OptionList.Content>
      </OptionList.Root>
      <HandleScope show={showPopup} setShow={setShowPopup} scope={activeTab} />
    </>
  );
};
export const handle = ({ workspace }: any) => {
  return {
    navbar: <WorkspaceTabs />,
    breadcrum: () => <CurrentBreadcrum {...{ workspace }} />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  const { workspace, scope } = ctx.params;

  ensureClusterSet(ctx);
  ensureAccountSet(ctx);

  const api =
    scope === 'workspace'
      ? GQLServerHandler(ctx.request).getWorkspace
      : GQLServerHandler(ctx.request).getEnvironment;

  try {
    const { data, errors } = await api({
      ...getScopeAndProjectQuery(ctx),
      name: workspace,
    });
    if (errors) {
      logger.error(errors);
      throw errors[0];
    }

    const { data: mTemplates, errors: mErrors } = await GQLServerHandler(
      ctx.request
    ).listTemplates({});
    if (mErrors) {
      throw mErrors[0];
    }

    return {
      workspace: data || {},
      managedTemplates: mTemplates || {},
    };
  } catch (err) {
    return redirect(
      `../../${scope === SCOPE.ENVIRONMENT ? 'environments' : 'workspaces'}`
    );
  }
};

export default Workspace;
