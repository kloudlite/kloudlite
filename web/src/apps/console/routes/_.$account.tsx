import { CaretDownFill, Plus } from '@jengaicons/react';
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
import { Button } from '~/components/atoms/button';
import OptionList from '~/components/atoms/option-list';
import Popup from '~/components/molecule/popup';
import logger from '~/root/lib/client/helpers/log';
import { useDataFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { IRemixCtx, LoaderResult } from '~/root/lib/types/common';
import {
  type IAccount,
  type IAccounts,
} from '../server/gql/queries/account-queries';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { parseName } from '../server/r-utils/common';
import { ensureAccountClientSide } from '../server/utils/auth-utils';
import { IConsoleRootContext } from './_/route';

// OptionList for various actions
const AccountMenu = ({ account }: { account: IAccount }) => {
  const accounts = useDataFromMatches<IAccounts>('accounts', {});
  const { account: accountName } = useParams();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Button
          content={account.displayName}
          variant="outline"
          suffix={<CaretDownFill />}
          size="sm"
        />
      </OptionList.Trigger>
      <OptionList.Content className="!pt-lg !pb-md">
        {accounts.map((acc) => {
          const name = parseName(acc);
          return (
            <OptionList.Item
              key={name}
              onClick={() => {
                if (accountName !== name) {
                  navigate(`/${name}/infra/clusters`);
                }
              }}
              active={accountName === name}
            >
              {/* <span>
                {name} {accountName === name ? '. active' : null}
              </span> */}
              {/* <div className="flex flex-col">
                <span className="bodyMd-medium text-text-default">
                  {acc.displayName}
                </span>
                <span className="bodySm text-text-soft">{name}</span>
              </div> */}
              <span title={name}>{acc.displayName}</span>
            </OptionList.Item>
          );
        })}
        <OptionList.Separator />
        <OptionList.Link to="/new-team" className="text-text-primary">
          <span className="text-icon-primary">
            <Plus size={16} />
          </span>
          <span>Create new team</span>
        </OptionList.Link>
      </OptionList.Content>
    </OptionList.Root>
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

export const handle = ({ account }: any) => {
  return {
    accountMenu: <AccountMenu account={account} />,
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

    acccountData = data;

    return {
      account: data,
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
