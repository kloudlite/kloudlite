import {
  CaretDownFill,
  ChevronUpDown,
  Plus,
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
import { Button } from '~/components/atoms/button';
import HandleConsoleDevices from '~/console/page-components/handle-console-devices';
import Profile from '~/components/molecule/profile';
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
  const { account } = useLoaderData();
  const rootContext = useOutletContext<IConsoleRootContext>();
  const { unloadState, reset, proceed } = useUnsavedChanges();

  const params = useParams();
  useEffect(() => {
    ensureAccountClientSide(params);
  }, []);
  return (
    <>
      <Outlet context={{ ...rootContext, account }} />
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
  const d = devices;

  const [visible, setVisible] = useState(false);
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
  return (
    devices?.length > 0 && (
      <OptionList.Root>
        <OptionList.Trigger>
          <Button
            content={devices?.[0].displayName}
            variant="outline"
            prefix={<WireGuardlogo />}
          />

          {/* <Button
            prefix={<WireGuardlogo />}
            content={devices?.[0].displayName || ''}
          /> */}
        </OptionList.Trigger>
        <OptionList.Content>
          <OptionList.Item>
            <div className="flex flex-col">
              <span className="bodyMd-medium text-text-default">
                {devices?.[0].displayName}
              </span>
              <span className="bodySm text-text-soft">test todo</span>
            </div>
          </OptionList.Item>
        </OptionList.Content>
      </OptionList.Root>
    )
  );
};

export const handle = ({ account, consoleVpnDevices }: any) => {
  console.log(consoleVpnDevices);

  return {
    breadcrum: () => <AccountMenu account={account} />,
    devicesMenu: () => <DevicesMenu devices={consoleVpnDevices} />,
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

    const { data: consoleVpnDevices, errors: vpnError } =
      await GQLServerHandler(ctx.request).listConsoleVpnDevicesForUser({});
    if (vpnError) {
      throw vpnError[0];
    }

    acccountData = data;
    return {
      account: data,
      consoleVpnDevices,
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
