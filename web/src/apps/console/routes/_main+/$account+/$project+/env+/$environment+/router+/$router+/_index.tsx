import { redirect } from '@remix-run/node';
import { IRemixCtx } from '~/lib/types/common';

export const loader = async (ctx: IRemixCtx) => {
  const { account, project, environment, router } = ctx.params;
  return redirect(`/${account}/${project}/env/${environment}/${router}/routes`);
};
