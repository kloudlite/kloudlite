import { redirect } from '@remix-run/node';

export const loader = async (ctx) => {
  const { account, project, cluster } = ctx.params;
  return redirect(`/${account}/${cluster}/${project}/workspaces`);
};
