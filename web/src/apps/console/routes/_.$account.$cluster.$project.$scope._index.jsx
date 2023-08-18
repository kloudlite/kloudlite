import { redirect } from '@remix-run/node';

export const loader = (ctx) => {
  const { project, account, cluster } = ctx.params;
  return redirect(`/${account}/${cluster}/${project}/workspaces`);
};
