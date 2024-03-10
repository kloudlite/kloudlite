import {
  BackingServices,
  CirclesFour,
  GearSix,
  Plus,
  Search,
  File,
  TreeStructure,
  Check,
  ChevronUpDown,
  ChevronDown,
} from '~/console/components/icons';
import { redirect } from '@remix-run/node';
import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { useRef, useState } from 'react';
import OptionList from '~/components/atoms/option-list';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import HandleScope from '~/console/page-components/new-scope';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
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
import { BreadcrumSlash, tabIconSize } from '~/console/utils/commons';
import {
  IEnvironment,
  IEnvironments,
} from '~/console/server/gql/queries/environment-queries';
import { cn } from '~/components/utils';
import { Button } from '~/components/atoms/button';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
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

const tabs = [
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <CirclesFour size={tabIconSize} />
        Apps
      </span>
    ),
    to: '/apps',
    value: '/apps',
  },
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <TreeStructure size={tabIconSize} />
        Router
      </span>
    ),
    to: '/routers',
    value: '/routers',
  },
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <File size={tabIconSize} />
        Configs and Secrets
      </span>
    ),
    to: '/cs/configs',
    value: '/cs',
  },
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <BackingServices size={tabIconSize} />
        Managed resources
      </span>
    ),
    to: '/managed-resources',
    value: '/managed-resources',
  },
  // {
  //   label: 'Jobs & Crons',
  //   to: '/jc/task',
  //   value: '/jc',
  // },
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <GearSix size={tabIconSize} />
        Settings
      </span>
    ),
    to: '/settings/general',
    value: '/settings',
  },
];

const EnvironmentTabs = () => {
  const { account, project, environment } = useParams();
  return (
    <CommonTabs baseurl={`/${account}/${project}/${environment}`} tabs={tabs} />
  );
};

const CurrentBreadcrum = ({ environment }: { environment: IEnvironment }) => {
  const params = useParams();

  const [showPopup, setShowPopup] = useState<any>(null);

  const api = useConsoleApi();
  const [search, setSearch] = useState('');
  const [searchText, setSearchText] = useState('');

  const { project, account } = params;

  const { data: environments, isLoading } = useCustomSwr(
    () => `/environments/${searchText}`,
    async () =>
      api.listEnvironments({
        search: {
          text: {
            matchType: 'regex',
            regex: searchText,
          },
        },
        projectName: project || '',
      })
  );

  useDebounce(
    () => {
      ensureAccountClientSide(params);
      setSearchText(search);
    },
    300,
    [search]
  );

  const [open, setOpen] = useState(false);

  return (
    <>
      <BreadcrumSlash />
      <span className="mx-md" />

      <OptionList.Root open={open} onOpenChange={setOpen} modal={false}>
        <OptionList.Trigger>
          <Button
            content={environment.displayName}
            size="sm"
            variant="plain"
            suffix={<ChevronDown />}
          />
        </OptionList.Trigger>
        <OptionList.Content className="!pt-0 !pb-md" align="center">
          <div className="p-[3px] pb-0">
            <OptionList.TextInput
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              prefixIcon={<Search />}
              focusRing={false}
              placeholder="Search environments"
              compact
              className="border-0 rounded-none"
            />
          </div>
          <OptionList.Separator />
          {parseNodes(environments)?.map((item) => {
            return (
              <OptionList.Link
                key={parseName(item)}
                LinkComponent={Link}
                to={`/${account}/${project}/${parseName(item)}`}
                className={cn(
                  'flex flex-row items-center justify-between',
                  parseName(item) === parseName(environment)
                    ? 'bg-surface-basic-pressed hover:!bg-surface-basic-pressed'
                    : ''
                )}
              >
                <span>{item.displayName}</span>
                {parseName(item) === parseName(environment) && (
                  <span>
                    <Check size={16} />
                  </span>
                )}
              </OptionList.Link>
            );
          })}

          {parseNodes(environments).length === 0 && !isLoading && (
            <div className="flex flex-col gap-lg max-w-[198px] px-xl py-lg">
              <div className="bodyLg-medium text-text-default">
                No environments found
              </div>
              <div className="bodyMd text-text-soft">
                Your search for "{search}" did not match and environments.
              </div>
            </div>
          )}

          {isLoading && parseNodes(environments).length === 0 && (
            <div className="min-h-7xl" />
          )}

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
