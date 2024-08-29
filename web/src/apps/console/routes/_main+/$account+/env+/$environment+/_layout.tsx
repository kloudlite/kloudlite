import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { useState } from 'react';
import Breadcrum from '~/console/components/breadcrum';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import {
  BackingServices,
  CirclesFour,
  File,
  GearSix,
} from '~/console/components/icons';
import HandleScope from '~/console/page-components/handle-environment';
import { ICluster } from '~/console/server/gql/queries/cluster-queries';
import { IEnvironment } from '~/console/server/gql/queries/environment-queries';
import { ILoginUrls, ILogins } from '~/console/server/gql/queries/git-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseName } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { BreadcrumSlash, tabIconSize } from '~/console/utils/commons';
import { SubNavDataProvider } from '~/lib/client/hooks/use-create-subnav-action';
import { IRemixCtx, LoaderResult } from '~/lib/types/common';
import logger from '~/root/lib/client/helpers/log';
import { handleError } from '~/root/lib/utils/common';
import { IAccountContext } from '../../_layout';

const Environment = () => {
  const rootContext = useOutletContext<IAccountContext>();
  const { environment, managedTemplates, loginUrls, logins, cluster } =
    useLoaderData();

  return (
    <SubNavDataProvider>
      <Outlet
        context={{
          ...rootContext,
          environment,
          managedTemplates,
          loginUrls,
          logins,
          cluster,
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
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <BackingServices size={tabIconSize} />
        Imported Managed Resources
      </span>
    ),
    to: '/managed-resources',
    value: '/managed-resources',
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
  // {
  //   label: (
  //     <span className="flex flex-row items-center gap-lg">
  //       <TreeStructure size={tabIconSize} />
  //       External Apps
  //     </span>
  //   ),
  //   to: '/external-apps',
  //   value: '/external-apps',
  // },
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
        label: 'Environments',
      }}
      baseurl={`/${account}/env/${environment}`}
      tabs={tabs}
    />
  );
};

const CurrentBreadcrum = ({ environment }: { environment: IEnvironment }) => {
  const params = useParams();

  const [showPopup, setShowPopup] = useState<any>(null);

  // const api = useConsoleApi();
  // const [search, setSearch] = useState('');
  // const [searchText, setSearchText] = useState('');

  const { account } = params;

  // const { data: environments, isLoading } = useCustomSwr(
  //   () => `/environments/${searchText}`,
  //   async () =>
  //     api.listEnvironments({
  //       search: {
  //         text: {
  //           matchType: 'regex',
  //           regex: searchText,
  //         },
  //       },
  //     })
  // );

  // useDebounce(
  //   () => {
  //     ensureAccountClientSide(params);
  //     setSearchText(search);
  //   },
  //   300,
  //   [search]
  // );

  // const [open, setOpen] = useState(false);
  // const buttonRef = useRef<HTMLButtonElement>(null);
  // const [isMouseOver, setIsMouseOver] = useState<boolean>(false);

  return (
    <>
      <BreadcrumSlash />
      <span className="mx-md" />
      <Breadcrum.Button
        to={`/${account}/environments`}
        linkComponent={Link}
        content="Environments"
      />
      <BreadcrumSlash />
      <span className="mx-md" />

      <Breadcrum.Button
        // prefix={
        //   <span className="p-md flex items-center justify-center rounded-full border border-border-default text-text-soft">
        //     <Buildings size={16} />
        //   </span>
        // }
        content={environment.displayName}
        size="sm"
        variant="plain"
        linkComponent={Link}
        to={`/${account}/env/${parseName(environment)}/apps`}
      />

      {/* <OptionList.Root open={open} onOpenChange={setOpen} modal={false}>
        <OptionList.Trigger>
          <button
            ref={buttonRef}
            aria-label="accounts"
            className={cn(
              'outline-none rounded py-lg px-md mx-md bg-surface-basic-hovered',
              open || isMouseOver ? 'bg-surface-basic-pressed' : ''
            )}
            onMouseOver={() => {
              setIsMouseOver(true);
            }}
            onMouseOut={() => {
              setIsMouseOver(false);
            }}
            onFocus={() => {
              //
            }}
            onBlur={() => {
              //
            }}
          >
            <div className="flex flex-row items-center gap-md">
              <ChevronUpDown size={16} />
            </div>
          </button>
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
          {parseNodes(environments)?.map((item, i) => {
            if (i > 5) {
              return null;
            }
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

          {parseNodes(environments).length > 5 && (
            <span className="bodySm-medium text-text-soft px-xl py-sm">
              search for more...
            </span>
          )}

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
      </OptionList.Root> */}
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

    envData = data;
    return {
      loginUrls,
      logins,
      environment: envData,
      cluster: cData,
    };
  } catch (err) {
    logger.error(err);
    return handleError(err) as {
      logins: ILogins;
      loginUrls: ILoginUrls;
      environment: IEnvironment;
      cluster: ICluster;
    };
  }
};

export interface IEnvironmentContext extends IAccountContext {
  logins: LoaderResult<typeof loader>['logins'];
  loginUrls: LoaderResult<typeof loader>['loginUrls'];
  environment: LoaderResult<typeof loader>['environment'];
  cluster: LoaderResult<typeof loader>['cluster'];
}

export default Environment;
