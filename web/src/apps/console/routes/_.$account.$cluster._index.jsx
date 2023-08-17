import { redirect } from '@remix-run/node';

export const loader = async (ctx) => {
  const { account, cluster } = ctx.params;
  return redirect(`/${account}/${cluster}/nodepools`);
};
