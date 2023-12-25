import { redirect } from '@remix-run/node';
import { IRemixCtx } from '~/root/lib/types/common';
import { GQLServerHandler } from '../server/gql/saved-queries';

export const loader = async (ctx: IRemixCtx) => {
  const { account } = ctx.params;
  const { errors } = await GQLServerHandler(ctx.request).getAccount({
    accountName: account,
  });
  if (errors) {
    return redirect('/teams');
  }

  return redirect(`/${account}/infra/clusters`);
};
