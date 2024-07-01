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
import Profile from '~/components/molecule/profile';
import { TopBar } from '~/components/organisms/top-bar';
import { generateKey, titleCase } from '~/components/utils';
import Breadcrum from '~/console/components/breadcrum';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import LogoWrapper from '~/console/components/logo-wrapper';
import { ViewModeProvider } from '~/console/components/view-mode';
import { IAccounts } from '~/console/server/gql/queries/account-queries';
import { setupAccountContext } from '~/console/server/utils/auth-utils';
import { constants } from '~/console/server/utils/constants';
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
import { IExtRemixCtx, IRemixCtx } from '~/root/lib/types/common';
import {
  InfraAsCode,
  GearSix,
  Project,
  BackingServices,
  BellFill,
  Sliders,
} from '~/console/components/icons';
import { Button, IconButton } from '~/components/atoms/button';
import { Avatar } from '~/components/atoms/avatar';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { ExtractNodeType, parseNodes } from '~/console/server/r-utils/common';
import { ICommsNotifications } from '~/console/server/gql/queries/comms-queries';
import { LoadingPlaceHolder } from '~/console/components/loading';

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

export const meta = (c: IRemixCtx) => {
  return [
    { title: `Account ${constants.metadot} ${c.params?.account || ''}` },
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
              Environments
            </span>
          ),
          to: '/environments',
          value: '/environments',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <BackingServices size={iconSize} />
              Integrated Services
            </span>
          ),
          to: '/managed-services',
          value: '/managed-services',
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
        // {
        //   label: (
        //     <span className="flex flex-row items-center gap-lg">
        //       <ContainerIcon size={iconSize} />
        //       Packages
        //     </span>
        //   ),
        //   to: '/packages/repos',
        //   value: '/packages',
        // },
        // {
        //   label: (
        //     <span className="flex flex-row items-center gap-lg">
        //       <WireGuardlogo size={iconSize} />
        //       Vpn Devices
        //     </span>
        //   ),
        //   to: '/vpn-devices',
        //   value: '/vpn-devices',
        // },
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
    <LogoWrapper to={`/${account}/environments`}>
      <BrandLogo />
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
const ProfileMenu = ({ hideProfileName }: { hideProfileName: boolean }) => {
  const { user } = useLoaderData();
  const cookie = getCookie();
  const { pathname } = useLocation();
  const eNavigate = useExternalRedirect();
  const { account } = useParams();

  return (
    <OptionList.Root>
      <OptionList.Trigger>
        <div>
          <div className="hidden md:flex">
            {!hideProfileName ? (
              <Profile name={titleCase(user.name)} size="xs" />
            ) : (
              <Profile size="xs" />
            )}
          </div>
          <div className="flex md:hidden">
            <Profile size="xs" />
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
        <OptionList.Link
          LinkComponent={Link}
          to={`/${account}/user-profile/account`}
        >
          Profile Settings
        </OptionList.Link>

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

// type INotificationMessage = {
//   id: string;
//   name: string;
//   message: string;
//   time: string;
//   isRead: boolean;
//   isInvited: boolean;
// };

type INotificationBaseType = ExtractNodeType<ICommsNotifications>;

const NotificationMessageView = ({
  notificationMessage,
}: {
  // notificationMessage: INotificationMessage;
  notificationMessage: INotificationBaseType;
}) => {
  const avatar = notificationMessage.read ? (
    <Avatar size="xs" />
  ) : (
    <Avatar size="xs" dot />
  );

  console.log('kkk', notificationMessage);

  return (
    <div className="flex flex-row gap-lg">
      {avatar}
      <div className="flex flex-col gap-xl">
        <div className="flex flex-col gap-md">
          <span className="flex">
            <span className="bodySm-medium">
              {notificationMessage.accountName}&nbsp;
            </span>
            <span className="bodySm text-text-soft">
              {notificationMessage.content.title}
            </span>
          </span>
          <span className="bodySm text-text-disabled">
            {notificationMessage.creationTime}
          </span>
        </div>
        {/* {notificationMessage.isInvited && (
          <BottomNavigation
            secondaryButton={{
              variant: 'outline',
              content: 'Decline',
              prefix: undefined,
              size: 'sm',
              // onClick: () => {
              //   navigate(`/${accountName}/environments`);
              // },
            }}
            primaryButton={{
              variant: 'primary',
              content: 'Acceept',
              // loading: isLoading,
              type: 'submit',
              size: 'sm',
            }}
          />
        )} */}
      </div>
    </div>
  );
};

const NotificationMenu = () => {
  // const { user } = useLoaderData();
  // const cookie = getCookie();
  // const { pathname } = useLocation();
  // const eNavigate = useExternalRedirect();
  // const { account } = useParams();
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { data: notificationsData, isLoading: notificationIsLoading } =
    useCustomSwr(
      'notifications',
      async () =>
        api.listNotifications({
          pagination: {
            first: 100,
          },
        }),
      true
    );

  const notifications = parseNodes(notificationsData);

  console.log('notifications', notificationsData);
  console.log('nnn', notifications);

  // const notificationMessage: INotificationMessage[] = [
  //   {
  //     id: '1',
  //     name: 'Piyush',
  //     message: 'invited you to the team kloudlite ops',
  //     time: '10 hrs ago',
  //     isRead: false,
  //     isInvited: true,
  //   },
  //   {
  //     id: '2',
  //     name: 'Bikash',
  //     message: 'deployed the application nginx',
  //     time: '10 hrs ago',
  //     isRead: true,
  //     isInvited: false,
  //   },
  //   {
  //     id: '3',
  //     name: 'Piyush',
  //     message: 'invited you to the team kloudlite ops',
  //     time: '10 hrs ago',
  //     isRead: false,
  //     isInvited: true,
  //   },
  //   {
  //     id: '4',
  //     name: 'Bikash',
  //     message: 'deployed the application nginx',
  //     time: '10 hrs ago',
  //     isRead: true,
  //     isInvited: false,
  //   },
  //   {
  //     id: '5',
  //     name: 'Piyush',
  //     message: 'invited you to the team kloudlite ops',
  //     time: '10 hrs ago',
  //     isRead: false,
  //     isInvited: true,
  //   },
  //   {
  //     id: '6',
  //     name: 'Bikash',
  //     message: 'deployed the application nginx',
  //     time: '10 hrs ago',
  //     isRead: true,
  //     isInvited: false,
  //   },
  //   {
  //     id: '7',
  //     name: 'Piyush',
  //     message: 'invited you to the team kloudlite ops',
  //     time: '10 hrs ago',
  //     isRead: false,
  //     isInvited: true,
  //   },
  //   {
  //     id: '8',
  //     name: 'Bikash',
  //     message: 'deployed the application nginx',
  //     time: '10 hrs ago',
  //     isRead: true,
  //     isInvited: false,
  //   },
  // ];

  return (
    <OptionList.Root>
      <OptionList.Trigger>
        <IconButton icon={<BellFill />} variant="plain" />
      </OptionList.Trigger>
      <OptionList.Content className="w-[360px] !max-w-[360px] !py-0 ">
        <div className="flex flex-row items-center justify-between p-2xl bg-surface-basic-active">
          <span className="headingMd">Notifications</span>
          <div className="flex flex-row">
            <Button
              size="sm"
              content={
                <span className="truncate text-left">Mark all as read</span>
              }
              variant="primary-plain"
              className="truncate"
              onClick={async () => {
                try {
                  const { errors: e } = await api.markAllNotificationAsRead();
                  if (e) {
                    throw e[0];
                  }
                  reloadPage();
                } catch (error) {
                  console.log(error);
                }
              }}
            />
            <IconButton icon={<Sliders />} variant="plain" />
          </div>
        </div>
        <div className="flex flex-col gap-3xl p-3xl max-h-[425px] overflow-y-scroll">
          {notificationIsLoading && <LoadingPlaceHolder />}
          {notifications.length === 0 ? (
            <div className="flex items-center justify-center bodyMd-medium text-text-soft">
              You dont have any notifications yet
            </div>
          ) : (
            notifications.map((data) => {
              return (
                <NotificationMessageView
                  key={data.id}
                  notificationMessage={data}
                />
              );
            })
          )}
          {/* {notifications.map((data) => {
            return (
              <NotificationMessageView
                key={data.id}
                notificationMessage={data}
              />
            );
          })} */}
        </div>
      </OptionList.Content>
    </OptionList.Root>
  );
};

const Console = () => {
  const loaderData = useLoaderData<typeof loader>();

  const matches = useMatches();

  const navbar = useHandleFromMatches('navbar', null);
  const logo = useHandleFromMatches('logo', null);

  const noMainLayout = useHandleFromMatches('noMainLayout', null);

  const devicesMenu = useHandleFromMatches('devicesMenu', null);
  const noBreadCrum = useHandleFromMatches('noBreadCrum', false);
  const hideProfileName = useHandleFromMatches('hideProfileName', false);

  const headerExtra = useHandleFromMatches('headerExtra', null);

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
          noBreadCrum ? null : (
            <Breadcrum.Root>
              {breadcrum.map((bc: any, index) =>
                cloneElement(bc.handle.breadcrum(bc), {
                  key: generateKey(index),
                })
              )}
            </Breadcrum.Root>
          )
        }
        logo={logo ? cloneElement(logo, { size: 24 }) : null}
        // tabs={navbar === constants.nan ? null : navbar}
        tabs={navbar === constants.nan ? null : navbar}
        actions={
          <div className="flex flex-row gap-2xl items-center">
            {!!devicesMenu && devicesMenu()}
            {!!headerExtra && headerExtra()}
            <NotificationMenu />
            <ProfileMenu hideProfileName={hideProfileName} />
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
