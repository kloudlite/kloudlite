import { redirect } from '@remix-run/node';

export const loader = async (ctx) => {
  const { account } = ctx.params;
  console.log(account);
  return redirect(`/${account}/projects`);
};
