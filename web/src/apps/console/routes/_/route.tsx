import {
  Link,
  Outlet,
  ShouldRevalidateFunction,
  useLoaderData,
  useLocation,
  useParams,
} from '@remix-run/react';
import { cloneElement, useCallback } from 'react';
import Container from '~/components/atoms/container';
import OptionList from '~/components/atoms/option-list';
import { BrandLogo } from '~/components/branding/brand-logo';
import { Profile } from '~/components/molecule/profile';
import { TopBar } from '~/components/organisms/top-bar';
import Breadcrum from '~/console/components/breadcrum';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { ViewModeProvider } from '~/console/components/view-mode';
import { IAccounts } from '~/console/server/gql/queries/account-queries';
import { setupAccountContext } from '~/console/server/utils/auth-utils';
import { constants } from '~/console/server/utils/constants';
import { LightTitlebarColor } from '~/design-system/tailwind-base';
import { getCookie } from '~/root/lib/app-setup/cookies';
import withContext from '~/root/lib/app-setup/with-contxt';
import { useExternalRedirect } from '~/root/lib/client/helpers/use-redirect';
import useMatches, {
  useHandleFromMatches,
} from '~/root/lib/client/hooks/use-custom-matches';
import { authBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { UserMe } from '~/root/lib/server/gql/saved-queries';
import { IExtRemixCtx } from '~/root/lib/types/common';

const restActions = (ctx: IExtRemixCtx) => {
  return withContext(ctx, {});
};

export const loader = async (ctx: IExtRemixCtx) => {
  return (await setupAccountContext(ctx)) || restActions(ctx);
};

export type IConsoleRootContext = {
  user: UserMe;
  accounts: IAccounts;
};

export const meta = () => {
  return [
    { title: 'Projects' },
    { name: 'theme-color', content: LightTitlebarColor },
  ];
};

const AccountTabs = () => {
  const { account } = useParams();
  return (
    <CommonTabs
      baseurl={`/${account}`}
      tabs={[
        {
          label: 'Projects',
          to: '/projects',
          value: '/projects',
        },
        {
          label: 'Clusters',
          to: '/clusters',
          value: '/clusters',
        },
        {
          label: 'Cloud providers',
          to: '/cloud-providers',
          value: '/cloud-providers',
        },
        {
          label: 'Domains',
          to: '/domains',
          value: '/domains',
        },
        // {
        //   label: 'Container registry',
        //   to: '/container-registry',
        //   value: '/container-registry',
        // },
        {
          label: 'VPN',
          to: '/vpn',
          value: '/vpn',
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

export const handle = () => {
  return {
    navbar: <AccountTabs />,
    logo: <BrandLogo detailed />,
  };
};

// OptionList for various actions
const ProfileMenu = () => {
  const { user } = useLoaderData();
  const cookie = getCookie();
  const { pathname } = useLocation();
  const eNavigate = useExternalRedirect();
  return (
    <OptionList.Root>
      <OptionList.Trigger>
        <div>
          <div className="hidden md:flex">
            <Profile name={user.name} size="xs" />
          </div>
          <div className="flex md:hidden">
            <Profile name={user.name} size="xs" />
          </div>
        </div>
      </OptionList.Trigger>
      <OptionList.Content className="w-[200px]">
        <OptionList.Item>
          <div className="flex flex-col">
            <span className="bodyMd-medium text-text-default">{user.name}</span>
            <span className="bodySm text-text-soft">{user.email}</span>
          </div>
        </OptionList.Item>
        <OptionList.Item>Profile Settings</OptionList.Item>
        <OptionList.Item>Manage account</OptionList.Item>
        <OptionList.Item>Notifications</OptionList.Item>
        <OptionList.Item>Support</OptionList.Item>
        <OptionList.Separator />
        <OptionList.Item
          onSelect={() => {
            cookie.set('url_history', pathname);
            eNavigate(`${authBaseUrl}/logout`);
          }}
        >
          Sign Out
        </OptionList.Item>
      </OptionList.Content>
    </OptionList.Root>
  );
};

const Console = () => {
  const loaderData = useLoaderData<typeof loader>();
  // const rootContext = useOutletContext();

  const { account: accountName } = useParams();

  const matches = useMatches();

  const navbar = useHandleFromMatches('navbar', null);
  const logo = useHandleFromMatches('logo', null);

  const noMainLayout = useHandleFromMatches('noMainLayout', null);

  const accountMenu = useHandleFromMatches('accountMenu', null);

  const breadcrum = useCallback(() => {
    return matches.filter((m) => m.handle?.breadcrum);
  }, [matches])();

  if (noMainLayout) {
    return (
      <Outlet
        context={{
          ...loaderData,
        }}
      />
    );
  }

  return (
    <div className="flex flex-col bg-surface-basic-subdued min-h-full">
      <TopBar
        fixed
        breadcrum={
          <Breadcrum.Root>
            {breadcrum.map((bc: any, index) =>
              cloneElement(bc.handle.breadcrum(bc), {
                key: index,
              })
            )}
          </Breadcrum.Root>
        }
        logo={
          <Link to={`/${accountName}/projects`} prefetch="intent">
            {logo ? cloneElement(logo, { size: 24 }) : null}
          </Link>
        }
        // tabs={navbar === constants.nan ? null : navbar}
        tabs={navbar === constants.nan ? null : navbar}
        actions={
          <div className="flex flex-row gap-2xl items-center">
            {!!accountMenu && accountMenu}
            <ProfileMenu />
          </div>
        }
      />
      <ViewModeProvider>
        <Container className="pb-5xl">
          <Outlet
            context={{
              ...loaderData,
            }}
          />
        </Container>
      </ViewModeProvider>
    </div>
  );
};

export const shouldRevalidate: ShouldRevalidateFunction = ({
  currentUrl,
  nextUrl,
  defaultShouldRevalidate,
}) => {
  if (!defaultShouldRevalidate) {
    return false;
  }
  if (currentUrl.search !== nextUrl.search) {
    return false;
  }
  return true;
};

export default Console;
