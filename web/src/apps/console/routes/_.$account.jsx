import { redirect } from '@remix-run/node';
import {
  Outlet,
  useOutletContext,
  useLoaderData,
  useNavigate,
  useParams,
} from '@remix-run/react';
import OptionList from '~/components/atoms/option-list';
import { Button } from '~/components/atoms/button';
import { Buildings, CaretDownFill, Plus } from '@jengaicons/react';
import { useState } from 'react';
import { GQLServerHandler } from '../server/gql/saved-queries';

// OptionList for various actions
const AccountMenu = ({ account, accounts }) => {
  const { account: accountName } = useParams();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
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
                if (accountName !== name) {
                  navigate(`/${name}/projects`);
                }
              }}
            >
              <Buildings size={16} />
              <span>
                {name} {accountName === name ? '. active' : null}
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
    </OptionList.Root>
  );
};

const Account = () => {
  const { account } = useLoaderData();
  const rootContext = useOutletContext();
  // @ts-ignore
  return <Outlet context={{ ...rootContext, account }} />;
};
export default Account;

export const handle = ({ account }) => {
  return {
    accountMenu: ({ accounts }) => (
      <>
        <AccountMenu account={account} accounts={accounts} />
        <div className="h-[15px] w-xs bg-border-default" />
      </>
    ),
  };
};

export const loader = async (ctx) => {
  const { account } = ctx.params;
  const { data, errors } = await GQLServerHandler(ctx.request).getAccount({
    accountName: account,
  });
  if (errors) {
    return redirect('/accounts');
  }
  return {
    account: data,
  };
};

export const shouldRevalidate = ({ currentUrl, nextUrl }) => {
  if (currentUrl.search !== nextUrl.search) {
    return false;
  }
  return true;
};
