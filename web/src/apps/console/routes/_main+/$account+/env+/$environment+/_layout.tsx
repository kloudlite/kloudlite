import {
  BackingServices,
  CirclesFour,
  GearSix,
  Plus,
  Search,
  File,
  Check,
  ChevronDown,
  Globe,
  ShieldCheck,
} from '~/console/components/icons';
import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { useState } from 'react';
import OptionList from '~/components/atoms/option-list';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import HandleScope from '~/console/page-components/new-scope';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseName, parseNodes } from '~/console/server/r-utils/common';
import {
  ensureAccountClientSide,
  ensureAccountSet,
} from '~/console/server/utils/auth-utils';
import { SubNavDataProvider } from '~/lib/client/hooks/use-create-subnav-action';
import useDebounce from '~/lib/client/hooks/use-debounce';
import { IRemixCtx, LoaderResult } from '~/lib/types/common';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { BreadcrumSlash, tabIconSize } from '~/console/utils/commons';
import { IEnvironment } from '~/console/server/gql/queries/environment-queries';
import { cn } from '~/components/utils';
import { Button } from '~/components/atoms/button';
import useCustomSwr from '~/lib/client/hooks/use-custom-swr';
import { ILoginUrls, ILogins } from '~/console/server/gql/queries/git-queries';
import logger from '~/root/lib/client/helpers/log';
import { IMSvTemplates } from '~/console/server/gql/queries/managed-templates-queries';
import { IAccountContext } from '../../_layout';

const Environment = () => {
  const rootContext = useOutletContext<IAccountContext>();
  const { environment, managedTemplates, loginUrls, logins, msvtemplates } =
    useLoaderData();

  return (
    <SubNavDataProvider>
      <Outlet
        context={{
          ...rootContext,
          msvtemplates,
          environment,
          managedTemplates,
          loginUrls,
          logins,
        }}
      />
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
  // {
  //   label: (
  //     <span className="flex flex-row items-center gap-lg">
  //       <TreeStructure size={tabIconSize} />
  //       Router
  //     </span>
  //   ),
  //   to: '/routers',
  //   value: '/routers',
  // },
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
  const { account, environment } = useParams();
  return (
    <CommonTabs
      backButton={{
        to: `/${account}/environments`,
        label: 'Envs',
      }}
      baseurl={`/${account}/env/${environment}`}
      tabs={tabs}
    />
  );
};

const CurrentBreadcrum = ({ environment }: { environment: IEnvironment }) => {
  const params = useParams();

  const [showPopup, setShowPopup] = useState<any>(null);

  const api = useConsoleApi();
  const [search, setSearch] = useState('');
  const [searchText, setSearchText] = useState('');

  const { account } = params;

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
            content={`${environment.displayName}`}
            size="sm"
            variant="plain"
            suffix={<ChevronDown />}
            prefix={
              environment.spec?.routing?.mode === 'private' ? (
                <ShieldCheck />
              ) : (
                <Globe />
              )
            }
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
                to={`/${account}/env/${parseName(item)}`}
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
                Your search for {`"${search}"`} did not match and environments.
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
  const { environment } = ctx.params;
  ensureAccountSet(ctx);

  let envData: IEnvironment;

  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getEnvironment(
      {
        name: environment,
      }
    );

    if (errors) {
      throw errors[0];
    }

    const { data: cData, errors: cErrors } = await GQLServerHandler(
      ctx.request
    ).getCluster({
      name: data.clusterName,
    });

    if (cErrors) {
      throw cErrors[0];
    }

    const { data: logins, errors: loginErrors } = await GQLServerHandler(
      ctx.request
    ).getLogins({});

    if (loginErrors) {
      throw loginErrors[0];
    }

    const { data: loginUrls, errors: dErrors } = await GQLServerHandler(
      ctx.request
    ).loginUrls({});

    if (dErrors) {
      throw dErrors[0];
    }

    const { data: msvTemplates, errors: msvError } = await GQLServerHandler(
      ctx.request
    ).listMSvTemplates({});

    if (msvError) {
      throw msvError[0];
    }

    envData = data;
    return {
      loginUrls,
      logins,
      environment: envData,
      cluster: cData,
      msvtemplates: msvTemplates || {},
    };
  } catch (err) {
    logger.error(err);

    const k: any = {};

    return {
      logins: k as ILogins,
      loginUrls: k as ILoginUrls,
      environment: k as IEnvironment,
      cluster: k as any,
    };
  }
};

export interface IEnvironmentContext extends IAccountContext {
  logins: LoaderResult<typeof loader>['logins'];
  loginUrls: LoaderResult<typeof loader>['loginUrls'];
  environment: LoaderResult<typeof loader>['environment'];
  cluster: LoaderResult<typeof loader>['cluster'];
  msvtemplates: IMSvTemplates;
}

export default Environment;
