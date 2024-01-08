import {
  ChevronDown,
  Plus,
  Search,
  VirtualMachine,
  Database,
} from '@jengaicons/react';
import { redirect } from '@remix-run/node';
import {
  Outlet,
  useLoaderData,
  useNavigate,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { useState } from 'react';
import OptionList from '~/components/atoms/option-list';
import Breadcrum from '~/console/components/breadcrum';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import HandleScope from '~/console/page-components/new-scope';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseName, parseNodes } from '~/console/server/r-utils/common';
import {
  ensureAccountClientSide,
  ensureAccountSet,
  ensureClusterClientSide,
} from '~/console/server/utils/auth-utils';
import logger from '~/root/lib/client/helpers/log';
import { SubNavDataProvider } from '~/root/lib/client/hooks/use-create-subnav-action';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { IRemixCtx } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import {
  BreadcrumButtonContent,
  BreadcrumSlash,
} from '~/console/utils/commons';
import MenuSelect from '~/console/components/menu-select';
import { IEnvironment } from '~/console/server/gql/queries/environment-queries';
import { toast } from '~/components/molecule/toast';
import { IProjectContext } from '../_layout';

export interface IEnvironmentContext extends IProjectContext {
  environment: IEnvironment;
}

const Environment = () => {
  const rootContext = useOutletContext<IProjectContext>();
  const { environment, managedTemplates } = useLoaderData();

  return (
    <SubNavDataProvider>
      <Outlet context={{ ...rootContext, environment, managedTemplates }} />
    </SubNavDataProvider>
  );
};

const EnvironmentTabs = () => {
  const { account, project, environment } = useParams();
  return (
    <CommonTabs
      baseurl={`/${account}/${project}/${environment}`}
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
          label: 'Settings',
          to: '/settings/general',
          value: '/settings',
        },
      ]}
    />
  );
};

const EnvironmentDropdown = () => {
  const navigate = useNavigate();
  const { account, project, cluster } = useParams();
  const iconSize = 14;

  const menuItems = [
    {
      label: (
        <span className="flex flex-row items-center gap-lg">
          <VirtualMachine size={iconSize} />
          Environments
        </span>
      ),
      value: `/${account}/${cluster}/${project}/environments`,
    },
    {
      label: (
        <span className="flex flex-row items-center gap-lg">
          <Database size={iconSize} />
          Managed Services
        </span>
      ),
      value: `/${account}/${cluster}/${project}/managed-services`,
    },
  ];
  return (
    <MenuSelect
      items={menuItems}
      value={`/${account}/${cluster}/${project}/environments`}
      onChange={(value) => navigate(value)}
      trigger={
        <Breadcrum.Button
          content={<BreadcrumButtonContent content="Environments" />}
        />
      }
    />
  );
};

// @ts-ignore
const CurrentBreadcrum = ({ environment }: { environment: IEnvironment }) => {
  const params = useParams();

  const [showPopup, setShowPopup] = useState<any>(null);
  const [environments, setEnvironments] = useState<IEnvironment[]>([]);

  const api = useConsoleApi();
  const [search, setSearch] = useState('');

  const { project } = params;

  useDebounce(
    async () => {
      ensureClusterClientSide(params);
      ensureAccountClientSide(params);
      if (!project) {
        throw new Error('Project is required.!');
      }
      try {
        const { data, errors } = await api.listEnvironments({
          projectName: project,
        });
        if (errors) {
          throw errors[0];
        }
        console.log(data);

        setEnvironments(parseNodes(data));
      } catch (err) {
        handleError(err);
      }
    },
    300,
    [search]
  );

  // const navigate = useNavigate();

  return (
    <>
      <BreadcrumSlash />
      {/* <EnvironmentDropdown /> */}
      {/* <BreadcrumChevronRight /> */}
      <span className="mx-sm" />
      <OptionList.Root>
        <OptionList.Trigger>
          <Breadcrum.Button
            variant="plain"
            size="sm"
            content={
              <div className="flex flex-row items-center gap-md">
                <BreadcrumButtonContent
                  className="bodyMd-semibold"
                  content={environment.displayName}
                />
                <ChevronDown size={12} />
              </div>
            }
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

          {environments.map((item) => {
            return (
              <OptionList.Item
                onClick={() => toast.info('todo')}
                key={parseName(item)}
              >
                {item.displayName}
              </OptionList.Item>
            );
          })}

          <OptionList.Separator />
          <OptionList.Item
            className="text-text-primary"
            onClick={() => setShowPopup({ type: 'add' })}
          >
            <Plus size={16} /> <span>New Environment</span>
          </OptionList.Item>
        </OptionList.Content>
      </OptionList.Root>
      <HandleScope show={showPopup} setShow={setShowPopup} />
    </>
  );
};
export const handle = ({ environment }: any) => {
  return {
    navbar: <EnvironmentTabs />,
    breadcrum: () => <CurrentBreadcrum {...{ environment }} />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  const { environment, project } = ctx.params;
  ensureAccountSet(ctx);

  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getEnvironment(
      {
        name: environment,
        projectName: project,
      }
    );

    if (errors) {
      logger.error(errors);
      throw errors[0];
    }

    return {
      environment: data || {},
    };
  } catch (err) {
    return redirect(`../environments`);
  }
};

export default Environment;
