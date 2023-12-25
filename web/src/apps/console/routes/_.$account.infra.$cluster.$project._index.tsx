import { redirect } from '@remix-run/node';
import { IRemixCtx } from '~/root/lib/types/common';

export const loader = async (ctx: IRemixCtx) => {
  const { account, project, cluster } = ctx.params;
  return redirect(`/${account}/${cluster}/${project}/environments`);
};
