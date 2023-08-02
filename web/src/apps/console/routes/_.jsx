import {
  BellSimpleFill,
  Buildings,
  CaretDownFill,
  Plus,
  SignOut,
} from '@jengaicons/react';
import {
  Link,
  Outlet,
  useMatches,
  useLoaderData,
  useOutletContext,
  useParams,
  useNavigate,
  useLocation,
} from '@remix-run/react';
import { Button, IconButton } from '~/components/atoms/button';
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
import { setupConsoleContext } from '../server/utils/auth-utils';

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
    href: '#',
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
    href: '#',
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

const Console = () => {
  const loaderData = useLoaderData();
  const rootContext = useOutletContext();

  const { account: accountName } = useParams();

  const matches = useMatches();

  const match = matches[matches.findLastIndex((m) => m.handle?.navbar)];
  const navbarData = match?.handle?.navbar
    ? match.handle?.navbar
    : defaultNavItems;
  const basepath = match?.data?.baseurl
    ? match.data?.baseurl
    : `/${accountName}`;

  const { activePath } = useActivePath({ parent: basepath });

  return (
    <div className="flex flex-col">
      <TopBar
        linkComponent={Link}
        fixed
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
            <AccountMenu />
            <div className="h-[15px] w-xs bg-border-default" />
            <div className="flex flex-row gap-lg items-center justify-center">
              <IconButton icon={BellSimpleFill} variant="plain" />
              <ProfileMenu />
            </div>
          </div>
        }
      />
      <Container>
        <Outlet context={{ ...rootContext, ...loaderData }} />
      </Container>
    </div>
  );
};

// OptionList for various actions
const ProfileMenu = ({ open, setOpen }) => {
  const { user } = useLoaderData();
  const cookie = getCookie();
  const { pathname } = useLocation();
  return (
    <OptionList open={open} onOpenChange={setOpen}>
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
            window.location.replace(`${authBaseUrl}/logout`);
          }}
        >
          <SignOut size={16} />
          <span>Logout</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList>
  );
};

// OptionList for various actions
const AccountMenu = ({ open, setOpen }) => {
  const { accounts, account } = useLoaderData();
  const { account: accountName } = useParams();
  const navigate = useNavigate();
  return (
    <OptionList open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Button
          content={account.name}
          variant="outline"
          suffix={CaretDownFill}
          size="sm"
        />
      </OptionList.Trigger>
      <OptionList.Content>
        {(accounts || []).map(({ name }) => {
          return (
            <OptionList.Item
              key={name}
              onSelect={() => {
                if (accountName !== account.name) {
                  navigate(`/${account.name}/projects`);
                }
              }}
            >
              <Buildings size={16} />
              <span>
                {name} . {accountName === account.name ? 'active' : null}
              </span>
            </OptionList.Item>
          );
        })}
        <OptionList.Item
          onSelect={() => {
            navigate(`/new-account`);
          }}
        >
          <Plus size={16} />
          <span>new account</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList>
  );
};

const restActions = (ctx) => {
  return withContext(ctx, {});
};

export const loader = async (ctx) => {
  return (await setupConsoleContext(ctx)) || restActions(ctx);
};

export default Console;
