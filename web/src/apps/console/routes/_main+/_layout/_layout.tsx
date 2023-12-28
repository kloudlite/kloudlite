import {
  Outlet,
  ShouldRevalidateFunction,
  useLoaderData,
  useLocation,
  useParams,
} from '@remix-run/react';
import { SetStateAction, cloneElement, useCallback, useState } from 'react';
import Container from '~/components/atoms/container';
import OptionList from '~/components/atoms/option-list';
import { BrandLogo } from '~/components/branding/brand-logo';
import Profile from '~/components/molecule/profile';
import { TopBar } from '~/components/organisms/top-bar';
import { generateKey, titleCase } from '~/components/utils';
import Breadcrum from '~/console/components/breadcrum';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import LogoWrapper from '~/console/components/logo-wrapper';
import { IShowDialog } from '~/console/components/types.d';
import { ViewModeProvider } from '~/console/components/view-mode';
import { IAccounts } from '~/console/server/gql/queries/account-queries';
import { setupAccountContext } from '~/console/server/utils/auth-utils';
import { constants } from '~/console/server/utils/constants';
import { DIALOG_TYPE } from '~/console/utils/commons';
import { LightTitlebarColor } from '~/design-system/tailwind-base';
import { getCookie } from '~/root/lib/app-setup/cookies';
import withContext from '~/root/lib/app-setup/with-contxt';
import { useExternalRedirect } from '~/root/lib/client/helpers/use-redirect';
import { SubNavDataProvider } from '~/root/lib/client/hooks/use-create-subnav-action';
import useMatches, {
  useHandleFromMatches,
} from '~/root/lib/client/hooks/use-custom-matches';
import { UnsavedChangesProvider } from '~/root/lib/client/hooks/use-unsaved-changes';
import { authBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { UserMe } from '~/root/lib/server/gql/saved-queries';
import { IExtRemixCtx } from '~/root/lib/types/common';
import {
  InfraAsCode,
  Container as ContainerIcon,
  GearSix,
  Project,
} from '@jengaicons/react';
import HandleProfile from './handle-profile';

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
  const iconSize = 16;
  return (
    <CommonTabs
      baseurl={`/${account}`}
      tabs={[
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <Project size={iconSize} />
              Projects
            </span>
          ),
          to: '/projects',
          value: '/projects',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <InfraAsCode size={iconSize} />
              Infrastructure
            </span>
          ),
          to: '/infra',
          value: '/infra',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <ContainerIcon size={iconSize} />
              Packages
            </span>
          ),
          to: '/packages/repos',
          value: '/packages',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <GearSix size={iconSize} />
              Settings
            </span>
          ),
          to: '/settings',
          value: '/settings',
        },
      ]}
    />
  );
};

const Logo = () => {
  const { account } = useParams();
  return (
    <LogoWrapper to={`/${account}/infra/clusters`}>
      <BrandLogo detailed />
    </LogoWrapper>
  );
};

export const handle = () => {
  return {
    navbar: <AccountTabs />,
    logo: <Logo />,
  };
};

// OptionList for various actions
const ProfileMenu = ({
  setShowProfileDialog,
}: {
  setShowProfileDialog: React.Dispatch<
    SetStateAction<IShowDialog<UserMe | null>>
  >;
}) => {
  const { user } = useLoaderData();
  const cookie = getCookie();
  const { pathname } = useLocation();
  const eNavigate = useExternalRedirect();

  return (
    <OptionList.Root>
      <OptionList.Trigger>
        <div>
          <div className="hidden md:flex">
            <Profile name={titleCase(user.name)} size="xs" />
          </div>
          <div className="flex md:hidden">
            <Profile name={user.name} size="xs" />
          </div>
        </div>
      </OptionList.Trigger>
      <OptionList.Content className="w-[200px]">
        <OptionList.Item>
          <div className="flex flex-col">
            <span className="bodyMd-medium text-text-default">
              {titleCase(user.name)}
            </span>
            <span className="bodySm text-text-soft">{user.email}</span>
          </div>
        </OptionList.Item>
        <OptionList.Item
          onClick={() => {
            setShowProfileDialog({ type: DIALOG_TYPE.NONE, data: user });
          }}
        >
          Profile Settings
        </OptionList.Item>

        <OptionList.Item>Notifications</OptionList.Item>
        <OptionList.Item>Support</OptionList.Item>
        <OptionList.Separator />
        <OptionList.Item
          onClick={() => {
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
  const [showProfileDialog, setShowProfileDialog] =
    useState<IShowDialog<UserMe | null>>(null);

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
                key: generateKey(index),
              })
            )}
          </Breadcrum.Root>
        }
        logo={logo ? cloneElement(logo, { size: 24 }) : null}
        // tabs={navbar === constants.nan ? null : navbar}
        tabs={navbar === constants.nan ? null : navbar}
        actions={
          <div className="flex flex-row gap-2xl items-center">
            {!!accountMenu && accountMenu}
            <ProfileMenu setShowProfileDialog={setShowProfileDialog} />
          </div>
        }
      />
      <ViewModeProvider>
        <SubNavDataProvider>
          <UnsavedChangesProvider>
            <Container className="pb-5xl">
              <Outlet
                context={{
                  ...loaderData,
                }}
              />
            </Container>
          </UnsavedChangesProvider>
        </SubNavDataProvider>
      </ViewModeProvider>
      <HandleProfile show={showProfileDialog} setShow={setShowProfileDialog} />
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
