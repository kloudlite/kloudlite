import { redirect } from '@remix-run/node';
import { GQLServerHandler } from '~/root/lib/server/gql/saved-queries';
import { IRemixCtx } from '~/root/lib/types/common';

export const loader = async (ctx: IRemixCtx) => {
  const { data } = await GQLServerHandler(ctx.request).whoAmI();
  if (data && !data.approved) {
    return redirect(`/invite-code`);
  }
  return redirect('/teams');
};
