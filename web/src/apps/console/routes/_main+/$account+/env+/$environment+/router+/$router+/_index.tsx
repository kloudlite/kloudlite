import { redirect } from '@remix-run/node';
import { IRemixCtx } from '~/lib/types/common';

export const loader = async (ctx: IRemixCtx) => {
  const { account, environment, router } = ctx.params;
  return redirect(`/${account}/env/${environment}/${router}/routes`);
};
