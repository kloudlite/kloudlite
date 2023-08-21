import {
  Outlet,
  useOutletContext,
  useLoaderData,
  useParams,
  useNavigate,
} from '@remix-run/react';
import OptionList from '~/components/atoms/option-list';
import { ChevronDown, Plus, Search } from '@jengaicons/react';
import Breadcrum from '~/console/components/breadcrum';
import { useEffect, useState } from 'react';
import {
  BlackProdLogo,
  BlackWorkspaceLogo,
} from '~/console/components/commons';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import logger from '~/root/lib/client/helpers/log';
import {
  getScopeAndProjectQuery,
  parseDisplayname,
  parseName,
  parseNodes,
} from '~/console/server/r-urils/common';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { toast } from '~/components/molecule/toast';
import Skeleton from 'react-loading-skeleton';
import HandleScope from '~/console/page-components/new-scope';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
// import { HandlePopup } from './handle-wrkspc-env';

const Workspace = () => {
  const rootContext = useOutletContext();
  const { workspace } = useLoaderData();

  // @ts-ignore
  return <Outlet context={{ ...rootContext, workspace }} />;
};

export default Workspace;

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
          to: '/cs',
          value: '/cs',
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

export const handle = ({ workspace }) => {
  return {
    navbar: <WorkspaceTabs />,
    breadcrum: () => <CurrentBreadcrum {...{ workspace }} />,
  };
};

const CurrentBreadcrum = ({ workspace }) => {
  const params = useParams();
  const { account, cluster, project, scope } = params;

  const [showPopup, setShowPopup] = useState(null);
  const [activeTab, setActiveTab] = useState(scope);
  const [workspaces, setWorkspaces] = useState([]);
  const [environments, setEnvironments] = useState([]);

  const api = useAPIClient();
  const [search, setSearch] = useState('');

  const [isLoading, setIsLoading] = useState(false);

  useDebounce(
    async () => {
      const listApi =
        activeTab === 'environment' ? api.listEnvironments : api.listWorkspaces;
      try {
        setIsLoading(true);
        const { data, errors } = await listApi({
          project: getScopeAndProjectQuery({ params }).project,
        });
        if (errors) {
          throw errors[0];
        }

        if (activeTab === 'environment') {
          setEnvironments(parseNodes(data));
        } else {
          setWorkspaces(parseNodes(data));
        }
      } catch (err) {
        toast.error(err.message);
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
            content={parseDisplayname(workspace)}
            prefix={BlackProdLogo}
            suffix={ChevronDown}
          />
        </OptionList.Trigger>
        <OptionList.Content className="!pt-0 !pb-md" align="center">
          <div className="p-[3px] pb-0">
            <OptionList.TextInput
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              prefixIcon={Search}
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
            onChange={setActiveTab}
            // LinkComponent={Link}
          >
            <OptionList.Tabs.Tab
              prefix={BlackWorkspaceLogo}
              label="Workspaces"
              value="workspace"
            />
            <OptionList.Tabs.Tab
              prefix={BlackProdLogo}
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
                    onSelect={() => {
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
                    {parseDisplayname(item)}
                  </OptionList.Item>
                );
              }
            )
          )}

          <OptionList.Separator />
          <OptionList.Item
            className="text-text-primary"
            onSelect={() => setShowPopup({ type: activeTab })}
          >
            <Plus size={16} />{' '}
            <span>
              {activeTab === 'workspace' ? 'New Workspace' : 'New Environment'}
            </span>
          </OptionList.Item>
        </OptionList.Content>
      </OptionList.Root>
      <HandleScope show={showPopup} setShow={setShowPopup} />
    </>
  );
};

export const loader = async (ctx) => {
  const { account, cluster, project, workspace, scope } = ctx.params;

  const api =
    scope === 'workspace'
      ? GQLServerHandler(ctx.request).getWorkspace
      : GQLServerHandler(ctx.request).getEnvironment;

  const { data, errors } = await api({
    project: {
      value: project,
      type: 'name',
    },
    name: workspace,
  });
  if (errors) {
    logger.error(errors);
  }

  return {
    workspace: data || {},
    account,
    project,
    cluster,
    scope,
  };
};
