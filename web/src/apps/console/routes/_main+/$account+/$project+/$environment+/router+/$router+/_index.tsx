import { redirect } from '@remix-run/node';
import { IRemixCtx } from '~/root/lib/types/common';

export const loader = async (ctx: IRemixCtx) => {
  const { account, project, environment, router } = ctx.params;
  return redirect(`/${account}/${project}/${environment}/${router}/routes`);
};
