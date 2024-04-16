import { redirect } from '@remix-run/node';
import { IRemixCtx } from '~/root/lib/types/common';

export const loader = async (ctx: IRemixCtx) => {
  const { account, project, deployment } = ctx.params;
  return redirect(`/${account}/${project}/deployment/${deployment}/devices`);
};
