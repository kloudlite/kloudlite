import { redirect } from '@remix-run/node';
import { IRemixCtx } from '~/root/lib/types/common';

export const loader = (ctx: IRemixCtx) => {
  const { project, account, cluster } = ctx.params;
  return redirect(`/${account}/${cluster}/${project}/environments`);
};
