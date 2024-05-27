import { redirect } from '@remix-run/node';
import { Outlet, useLoaderData, useOutletContext } from '@remix-run/react';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { IRemixCtx } from '~/root/lib/types/common';

const Account = () => {
  const { account } = useLoaderData();
  const rootContext: object = useOutletContext();

  return <Outlet context={{ ...rootContext, account }} />;
};
export default Account;

export const loader = async (ctx: IRemixCtx) => {
  const { a: account } = ctx.params;
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
