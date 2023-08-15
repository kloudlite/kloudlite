import { BellSimpleFill, ChevronDown, SignOut } from '@jengaicons/react';
import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
  useLocation,
} from '@remix-run/react';
import { IconButton } from '~/components/atoms/button';
import Container from '~/components/atoms/container';
import OptionList from '~/components/atoms/option-list';
import { BrandLogo } from '~/components/branding/brand-logo';
import { Profile } from '~/components/molecule/profile';
import { TopBar } from '~/components/organisms/top-bar';
import { LightTitlebarColor } from '~/design-system/tailwind-base';
import withContext from '~/root/lib/app-setup/with-contxt';
import { useActivePath } from '~/root/lib/client/hooks/use-active-path';
import { authBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { getCookie } from '~/root/lib/app-setup/cookies';
import { useExternalRedirect } from '~/root/lib/client/helpers/use-redirect';
import useMatches from '~/root/lib/client/hooks/use-custom-matches';
import { useCallback, useState } from 'react';
import { ProdLogo } from '~/components/branding/prod-logo';
import { NavTabs } from '~/components/atoms/tabs';
import { setupAccountContext } from '../server/utils/auth-utils';
import * as Breadcrum from '../components/breadcrum';

export const meta = () => {
  return [
    { title: 'Projects' },
    { name: 'theme-color', content: LightTitlebarColor },
  ];
};

const defaultNavItems = [
  {
    label: 'Projects',
    href: '/projects',
    key: 'projects',
    value: '/projects',
  },
  {
    label: 'Clusters',
    href: '/clusters',
    key: 'clusters',
    value: '/clusters',
  },
  {
    label: 'Cloud providers',
    href: '/cloud-providers',
    key: 'cloud-providers',
    value: '/cloud-providers',
  },
  {
    label: 'Domains',
    href: '/domains',
    key: 'domains',
    value: '/domains',
  },
  {
    label: 'Container registry',
    href: '/container-registry',
    key: 'container-registry',
    value: '/container-registry',
  },
  {
    label: 'VPN',
    href: '/vpn',
    key: 'vpn',
    value: '/vpn',
  },
  {
    label: 'Settings',
    href: '/settings/general',
    key: 'settings',
    value: '/settings',
  },
];

const BlackProdLogo = () => {
  return <ProdLogo className="fill-icon-default" size={16} />;
};

const Console = () => {
  const loaderData = useLoaderData();
  const rootContext = useOutletContext();

  const { account: accountName } = useParams();

  const matches = useMatches();

  // const match = matches[matches.findLastIndex((m) => m.handle?.navbar)];

  const match = useCallback(() => {
    return matches.reverse().find((m) => m.handle?.navbar);
  }, [matches])();

  const navbarData = match?.handle?.navbar
    ? match.handle?.navbar
    : defaultNavItems;

  const basepath = match?.data?.baseurl
    ? match.data?.baseurl
    : `/${accountName}`;

  const { activePath } = useActivePath({ parent: basepath });

  const accountMenu = useCallback(() => {
    return matches.reverse().find((m) => m.handle?.accountMenu)?.handle
      .accountMenu;
  }, [matches])();

  return (
    <div className="flex flex-col bg-surface-basic-subdued h-full">
      <TopBar
        linkComponent={Link}
        fixed
        breadcrum={
          <Breadcrum.Breadcrum>
            <Breadcrum.Button content="Lobster Early" />
            <Test />
          </Breadcrum.Breadcrum>
        }
        logo={
          <div>
            <div className="hidden md:block">
              <BrandLogo detailed size={24} />
            </div>
            <div className="block md:hidden">
              <BrandLogo size={24} />
            </div>
          </div>
        }
        tab={{
          basePath: basepath,
          value: `/${activePath.split('/')[1]}`,
          fitted: true,
          layoutId: 'console',
          items: navbarData,
        }}
        actions={
          <div className="flex flex-row gap-2xl items-center">
            {/* <AccountMenu /> */}
            {accountMenu && accountMenu(loaderData)}
            <div className="flex flex-row gap-lg items-center justify-center">
              <IconButton icon={BellSimpleFill} variant="plain" />
              <ProfileMenu />
            </div>
          </div>
        }
      />
      <Container>
        <Outlet
          context={{
            // @ts-ignore
            ...rootContext,
            ...loaderData,
          }}
        />
      </Container>
    </div>
  );
};

const Test = ({ open, setOpen }) => {
  const [data, setData] = useState([
    { checked: false, content: 'Verified', id: 'verified' },
    { checked: false, content: 'Un-Verified', id: 'unverified' },
  ]);
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Breadcrum.Button content="button" />
      </OptionList.Trigger>
      <OptionList.Content compact>
        <OptionList.TextInput compact className="border-0 border-b" />
        <OptionList.Tabs />
      </OptionList.Content>
    </OptionList.Root>
  );
};

// OptionList for various actions
const ProfileMenu = ({ open = false, setOpen = (_) => _ }) => {
  const { user } = useLoaderData();
  const cookie = getCookie();
  const { pathname } = useLocation();
  const eNavigate = useExternalRedirect();
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
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
      <OptionList.Content>
        <OptionList.Item
          onSelect={() => {
            cookie.set('url_history', pathname);
            eNavigate(`${authBaseUrl}/logout`);
          }}
        >
          <SignOut size={16} />
          <span>Logout</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList.Root>
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
