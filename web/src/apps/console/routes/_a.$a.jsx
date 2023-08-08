import { redirect } from '@remix-run/node';
import { Outlet, useOutletContext, useLoaderData } from '@remix-run/react';
import { GQLServerHandler } from '../server/gql/saved-queries';

const Account = () => {
  const { account } = useLoaderData();
  const rootContext = useOutletContext();
  return <Outlet context={{ ...rootContext, account }} />;
};
export default Account;

export const loader = async (ctx) => {
  const { a: account } = ctx.params;
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
