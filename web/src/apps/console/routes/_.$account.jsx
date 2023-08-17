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

const switchRoute = (route) => {
  const canSwitch = false;
  switch (route) {
  }
};

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
      <AccountMenu account={account} accounts={accounts} />
    ),
  };
};

export const loader = async (ctx) => {
  const { account } = ctx.params;
  const { data, errors } = await GQLServerHandler(ctx.request).getAccount({
    accountName: account,
  });
  if (errors) {
    return redirect('/teams');
  }
  return {
    account: data,
  };
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
