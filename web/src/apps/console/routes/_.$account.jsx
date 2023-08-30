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
import { CaretDownFill, Plus } from '@jengaicons/react';
import { useState } from 'react';
import withContext from '~/root/lib/app-setup/with-contxt';
import logger from '~/root/lib/client/helpers/log';
import { useDataFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { parseDisplayname, parseName } from '../server/r-urils/common';

// OptionList for various actions
const AccountMenu = ({ account }) => {
  const accounts = useDataFromMatches('accounts', {});
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
        {(accounts || []).map((acc) => {
          const name = parseName(acc);
          const displayName = parseDisplayname(acc);
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
  return <Outlet context={{ ...rootContext, account }} />;
};
export default Account;

export const handle = ({ account }) => {
  return {
    accountMenu: <AccountMenu account={account} />,
  };
};

export const loader = async (ctx) => {
  const { account } = ctx.params;
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getAccount({
      accountName: account,
    });
    if (errors) {
      throw errors[0];
    }

    return withContext(ctx, {
      account: data,
    });
  } catch (err) {
    logger.error(err);
    return redirect('/teams');
  }
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
