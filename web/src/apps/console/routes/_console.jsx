import {
  BellSimpleFill,
  CaretDownFill,
  Gear,
  SignOut,
  User,
} from '@jengaicons/react';
import {
  Link,
  Outlet,
  useMatch,
  useLocation,
  useMatches,
  useLoaderData,
  useOutletContext,
} from '@remix-run/react';
import { Button, IconButton } from '~/components/atoms/button';
import Container from '~/components/atoms/container';
import OptionList from '~/components/atoms/option-list';
import { BrandLogo } from '~/components/branding/brand-logo';
import { Profile } from '~/components/molecule/profile';
import { TopBar } from '~/components/organisms/top-bar';
import { LightTitlebarColor } from '~/design-system/tailwind-base';
import withContext from '~/root/lib/app-setup/with-contxt';
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
    value: 'projects',
  },
  {
    label: 'Clusters',
    href: '/clusters',
    key: 'clusters',
    value: 'clusters',
  },
  {
    label: 'Cloud providers',
    href: '#',
    key: 'cloudproviders',
    value: 'cloudproviders',
  },
  {
    label: 'Domains',
    href: '#',
    key: 'domains',
    value: 'domains',
  },
  {
    label: 'Container registry',
    href: '#',
    value: 'containerregistry',
    key: 'containerregistry',
  },
  {
    label: 'VPN',
    href: '#',
    key: 'vpn',
    value: 'vpn',
  },
  {
    label: 'Settings',
    href: '/settings/general',
    key: 'settings',
    value: 'settings',
  },
];

const Console = () => {
  const location = useLocation();
  const match = useMatch(
    {
      path: '/:path/*',
    },

    location.pathname
  );

  const loaderData = useLoaderData();
  const rootContext = useOutletContext();

  const matches = useMatches();
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
          value: match?.params?.path,
          fitted: true,
          layoutId: 'console',
          items: matches[matches.findLastIndex((m) => m.handle?.navbar)]
            ? matches[matches.findLastIndex((m) => m.handle?.navbar)].handle
                .navbar
            : defaultNavItems,
        }}
        actions={
          <div className="flex flex-row gap-2xl items-center">
            <ProfileMenu />
            <div className="h-[15px] w-xs bg-border-default" />
            <div className="flex flex-row gap-lg items-center justify-center">
              <IconButton icon={BellSimpleFill} variant="plain" />
              <AccountMenu />
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
const AccountMenu = ({ open, setOpen }) => {
  return (
    <OptionList open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <div>
          <div className="hidden md:flex">
            <Profile name="Astroman" subtitle={null} />
          </div>
          <div className="flex md:hidden">
            <Profile name="Astroman" size="small" subtitle={null} />
          </div>
        </div>
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.Item>
          <SignOut size={16} />
          <span>Logout</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList>
  );
};

// OptionList for various actions
const ProfileMenu = ({ open, setOpen }) => {
  return (
    <OptionList open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Button content="Nuveo" variant="outline" suffix={CaretDownFill} />
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.Item>
          <User size={16} />
          <span>Astroman</span>
        </OptionList.Item>
        <OptionList.Item>
          <Gear size={16} />
          <span>Ironman</span>
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
