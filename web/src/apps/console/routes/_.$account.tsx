import { redirect } from '@remix-run/node';
import {
  Outlet,
  useOutletContext,
  useLoaderData,
  useNavigate,
  useParams,
  ShouldRevalidateFunction,
} from '@remix-run/react';
import OptionList from '~/components/atoms/option-list';
import { Button } from '~/components/atoms/button';
import { CaretDownFill, Plus } from '@jengaicons/react';
import { useState } from 'react';
import logger from '~/root/lib/client/helpers/log';
import { useDataFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
import { IRemixCtx } from '~/root/lib/types/common';
import { parseName } from '~/root/src/generated/r-types/utils';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { IConsoleRootContext } from './_';
import {
  type IAccount,
  type IAccounts,
} from '../server/gql/queries/account-queries';

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
          content={parseName(account)}
          variant="outline"
          suffix={<CaretDownFill />}
          size="sm"
        />
      </OptionList.Trigger>
      <OptionList.Content>
        {accounts.map((acc) => {
          const name = parseName(acc);
          return (
            <OptionList.Item
              key={name}
              onSelect={() => {
                if (accountName !== name) {
                  navigate(`/${name}/projects`);
                }
              }}
            >
              <span>
                {name} {accountName === name ? '. active' : null}
              </span>
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

export interface IAccountContext extends IConsoleRootContext {
  account: IAccount;
}

const Account = () => {
  const { account } = useLoaderData();
  const rootContext = useOutletContext<IConsoleRootContext>();

  return <Outlet context={{ ...rootContext, account }} />;
};

export const handle = ({ account }: any) => {
  return {
    accountMenu: <AccountMenu account={account} />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  const { account } = ctx.params;
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getAccount({
      accountName: account,
    });
    if (errors) {
      throw errors[0];
    }

    return {
      account: data,
    };
  } catch (err) {
    logger.error(err);
    return redirect('/teams');
  }
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

export default Account;
