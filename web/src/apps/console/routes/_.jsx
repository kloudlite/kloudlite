import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
  useLocation,
} from '@remix-run/react';
import Container from '~/components/atoms/container';
import OptionList from '~/components/atoms/option-list';
import { BrandLogo } from '~/components/branding/brand-logo';
import { Profile } from '~/components/molecule/profile';
import { TopBar } from '~/components/organisms/top-bar';
import { LightTitlebarColor } from '~/design-system/tailwind-base';
import withContext from '~/root/lib/app-setup/with-contxt';
import { authBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { getCookie } from '~/root/lib/app-setup/cookies';
import { useExternalRedirect } from '~/root/lib/client/helpers/use-redirect';
import useMatches, {
  useHandleFromMatches,
} from '~/root/lib/client/hooks/use-custom-matches';
import { cloneElement, useCallback } from 'react';
import { setupAccountContext } from '../server/utils/auth-utils';
import Breadcrum from '../components/breadcrum';
import { CommonTabs } from '../components/common-navbar-tabs';
import { constants } from '../server/utils/constants';

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
        {
          label: 'Container registry',
          to: '/container-registry',
          value: '/container-registry',
        },
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
            <Profile name={user.name} size="xs" subtitle={null} />
          </div>
          <div className="flex md:hidden">
            <Profile name={user.name} size="xs" subtitle={null} />
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
  const loaderData = useLoaderData();
  const rootContext = useOutletContext();

  const { account: accountName } = useParams();

  const matches = useMatches();

  const navbar = useHandleFromMatches('navbar', null);

  const accountMenu = useHandleFromMatches('accountMenu', null);

  const breadcrum = useCallback(() => {
    return matches.filter((m) => m.handle?.breadcrum);
  }, [matches])();

  return (
    <div className="flex flex-col bg-surface-basic-subdued h-full">
      <TopBar
        fixed
        breadcrum={
          <Breadcrum.Root>
            {breadcrum.map((bc, index) =>
              // eslint-disable-next-line react/no-array-index-key
              cloneElement(bc.handle.breadcrum(bc), {
                key: index,
              })
            )}
          </Breadcrum.Root>
        }
        logo={
          <Link to={`/${accountName}/projects`} prefetch="intent">
            <div className="hidden md:block">
              <BrandLogo detailed size={24} />
            </div>
            <div className="block md:hidden">
              <BrandLogo size={24} />
            </div>
          </Link>
        }
        tabs={navbar === constants.nan ? null : navbar}
        actions={
          <div className="flex flex-row gap-2xl items-center">
            {/* <AccountMenu /> */}
            {accountMenu && accountMenu(loaderData)}
            <ProfileMenu />
          </div>
        }
      />
      <Container className="pb-5xl">
        <Outlet
          context={{
            ...rootContext,
            ...loaderData,
          }}
        />
      </Container>
    </div>
  );
};

const restActions = (ctx) => {
  return withContext(ctx, {});
};

export const loader = async (ctx) => {
  return (await setupAccountContext(ctx)) || restActions(ctx);
};

export const shouldRevalidate = ({
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
