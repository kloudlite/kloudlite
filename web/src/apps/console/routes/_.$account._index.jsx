import { redirect } from '@remix-run/node';
import { GQLServerHandler } from '../server/gql/saved-queries';

export const loader = async (ctx) => {
  const { account } = ctx.params;
  const { errors } = await GQLServerHandler(ctx.request).getAccount({
    accountName: account,
  });
  if (errors) {
    return redirect('/teams');
  }

  return redirect(`/${account}/projects`);
};
