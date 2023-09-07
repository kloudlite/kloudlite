import { redirect } from '@remix-run/node';
import { IRemixCtx } from '~/root/lib/types/common';

export const loader = async (ctx: IRemixCtx) => {
  const { account, cluster } = ctx.params;
  return redirect(`/${account}/${cluster}/nodepools`);
};
