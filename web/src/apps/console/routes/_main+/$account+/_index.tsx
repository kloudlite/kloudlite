import { redirect } from '@remix-run/node';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { IRemixCtx } from '~/root/lib/types/common';

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
