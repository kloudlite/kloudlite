import { redirect } from '@remix-run/node';
import { Outlet } from '@remix-run/react';
import { GQLServerHandler } from '../server/gql/saved-queries';

const Account = () => {
  return <Outlet />;
};
export default Account;

export const loader = async (ctx) => {
  console.log('here');
  const { account } = ctx.params;
  const { errors } = await GQLServerHandler(ctx.request).getAccount({
    accountName: account,
  });
  if (errors) {
    return redirect('/accounts');
  }
  return {};
};
