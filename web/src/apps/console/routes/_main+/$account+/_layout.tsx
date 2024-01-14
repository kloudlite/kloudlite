import {
  ArrowCounterClockwise,
  ArrowsCounterClockwise,
  CaretDownFill,
  ChevronUpDown,
  Copy,
  GearSix,
  Plus,
  QrCode,
  WireGuardlogo,
} from '@jengaicons/react';
import { redirect } from '@remix-run/node';
import {
  Outlet,
  ShouldRevalidateFunction,
  useLoaderData,
  useNavigate,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { useEffect, useState } from 'react';
import Popup from '~/components/molecule/popup';
import logger from '~/root/lib/client/helpers/log';
import { useDataFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { IRemixCtx, LoaderResult } from '~/root/lib/types/common';

import {
  IAccount,
  IAccounts,
} from '~/console/server/gql/queries/account-queries';
import { parseName, parseNodes } from '~/console/server/r-utils/common';

import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import MenuSelect from '~/console/components/menu-select';
import { BreadcrumButtonContent } from '~/console/utils/commons';
import OptionList from '~/components/atoms/option-list';
import { IConsoleDevicesForUser } from '~/console/server/gql/queries/console-vpn-queries';
import { Button, IconButton } from '~/components/atoms/button';
import HandleConsoleDevices, {
  QRCodeView,
  ShowWireguardConfig,
  decodeConfig,
  switchEnvironment,
} from '~/console/page-components/handle-console-devices';
import Profile from '~/components/molecule/profile';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { IConsoleRootContext } from '../_layout/_layout';

const AccountMenu = ({ account }: { account: IAccount }) => {
  const accounts = useDataFromMatches<IAccounts>('accounts', {});
  const { account: accountName } = useParams();
  const navigate = useNavigate();
  return (
    <MenuSelect
      value={accountName}
      items={accounts.map((acc) => ({
        label: acc.displayName,
        value: parseName(acc),
      }))}
      onChange={(value) => {
        navigate(`/${value}/projects`);
      }}
      trigger={
        <span className="flex flex-row items-center gap-md bodyMd text-text-default px-md py-sm">
          <BreadcrumButtonContent content={account.displayName} />
          <span className="text-icon-disabled">
            <ChevronUpDown color="currentColor" size={11} />
          </span>
        </span>
      }
    />
  );
};

const Account = () => {
  const { account, devicesForUser } = useLoaderData();
  const rootContext = useOutletContext<IConsoleRootContext>();
  const { unloadState, reset, proceed } = useUnsavedChanges();

  const params = useParams();
  useEffect(() => {
    ensureAccountClientSide(params);
  }, []);
  return (
    <>
      <Outlet context={{ ...rootContext, account, devicesForUser }} />
      <Popup.Root
        show={unloadState === 'blocked'}
        onOpenChange={() => {
          reset?.();
        }}
      >
        <Popup.Header>Unsaved changes</Popup.Header>
        <Popup.Content>Are you sure you discard the changes?</Popup.Content>
        <Popup.Footer>
          <Popup.Button
            content="Cancel"
            variant="basic"
            onClick={() => reset?.()}
          />
          <Popup.Button
            content="Discard"
            variant="warning"
            onClick={() => proceed?.()}
          />
        </Popup.Footer>
      </Popup.Root>
    </>
  );
};

const DevicesMenu = ({ devices }: { devices: IConsoleDevicesForUser }) => {
  const [visible, setVisible] = useState(false);
  const [isUpdate, setIsUpdate] = useState(false);
  const [showQR, setShowQR] = useState(false);

  const { copy } = useClipboard({
    onSuccess: () => {
      toast.success('WG config copied successfully');
    },
  });
  const api = useConsoleApi();
  const reload = useReload();

  const { environment, project } = useParams();

  if (!devices || devices?.length === 0) {
    return (
      <div>
        <Button
          content="Create new device"
          variant="basic"
          suffix={<Plus />}
          onClick={() => {
            setVisible(true);
          }}
        />
        <HandleConsoleDevices
          {...{
            isUpdate: false,
            visible,
            setVisible: () => setVisible(false),
          }}
        />
      </div>
    );
  }

  const device = devices[0];

  const getConfig = async () => {
    try {
      const { errors, data: out } = await api.getConsoleVpnDevice({
        name: parseName(device),
      });
      if (errors) {
        throw errors[0];
      }
      if (out.wireguardConfig) copy(decodeConfig(out.wireguardConfig));
    } catch (error) {
      handleError(error);
    }
  };
  return (
    devices?.length > 0 && (
      <>
        <OptionList.Root>
          <OptionList.Trigger>
            <IconButton variant="outline" icon={<WireGuardlogo />} />
          </OptionList.Trigger>
          <OptionList.Content>
            {device.environmentName && device.projectName && (
              <>
                <OptionList.Item>
                  <div className="flex flex-row items-center gap-lg">
                    <div className="flex flex-col">
                      <span className="bodyMd-medium text-text-default">
                        {device.displayName}
                      </span>
                      <span className="bodySm text-text-soft">
                        ({parseName(device)})
                      </span>
                    </div>
                  </div>
                </OptionList.Item>
                <OptionList.Separator />
                <OptionList.Item>
                  <div className="flex flex-col">
                    <span className="bodyMd-medium text-text-default">
                      Connected
                    </span>
                    <span className="bodySm text-text-soft">
                      {device.projectName}/{device.environmentName}
                    </span>
                  </div>
                </OptionList.Item>
                <OptionList.Item
                  onClick={() => {
                    setShowQR(true);
                  }}
                >
                  <div className="flex flex-row items-center gap-lg">
                    <div>
                      <QrCode size={16} />
                    </div>
                    <div>Show QR Code</div>
                  </div>
                </OptionList.Item>
                <OptionList.Item
                  onClick={() => {
                    getConfig();
                  }}
                >
                  <div className="flex flex-row items-center gap-lg">
                    <div>
                      <Copy size={16} />
                    </div>
                    <div>Copy WG Config</div>
                  </div>
                  <OptionList.Separator />
                </OptionList.Item>
                {environment &&
                  project &&
                  environment !== device.environmentName && (
                    <OptionList.Item
                      onClick={async (e) => {
                        await switchEnvironment({
                          api,
                          device,
                          environment,
                          project,
                        });
                        reload();
                      }}
                    >
                      <div className="flex flex-row items-center gap-lg">
                        <div>
                          <ArrowsCounterClockwise size={16} />
                        </div>
                        <div>Switch to {environment}</div>
                      </div>
                    </OptionList.Item>
                  )}
              </>
            )}
            <OptionList.Separator />
            <OptionList.Item
              onClick={() => {
                setIsUpdate(true);
              }}
            >
              <div className="flex flex-row items-center gap-lg">
                <div>
                  <GearSix size={16} />
                </div>
                <div>Settings</div>
              </div>
            </OptionList.Item>
          </OptionList.Content>
        </OptionList.Root>
        <HandleConsoleDevices
          {...{
            isUpdate: true,
            data: device,
            visible: isUpdate,
            setVisible: () => setIsUpdate(false),
          }}
        />
        <ShowWireguardConfig
          {...{
            visible: showQR,
            setVisible: () => setShowQR(false),
            data: { device: parseName(device) },
            mode: 'qr',
          }}
        />
      </>
    )
  );
};

export const handle = ({ account, devicesForUser }: any) => {
  return {
    breadcrum: () => <AccountMenu account={account} />,
    devicesMenu: () => <DevicesMenu devices={devicesForUser} />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  const { account } = ctx.params;
  let acccountData: IAccount;

  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getAccount({
      accountName: account,
    });
    if (errors) {
      throw errors[0];
    }

    const { data: devicesForUser, errors: vpnError } = await GQLServerHandler(
      ctx.request
    ).listConsoleVpnDevicesForUser({});
    if (vpnError) {
      throw vpnError[0];
    }

    acccountData = data;
    return {
      account: data,
      devicesForUser,
    };
  } catch (err) {
    logger.error(err);
    const k = redirect('/teams') as any;
    return k as {
      account: typeof acccountData;
    };
  }
};

export interface IAccountContext extends IConsoleRootContext {
  account: LoaderResult<typeof loader>['account'];
  devicesForUser: IConsoleDevicesForUser;
}

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

export default Account;
