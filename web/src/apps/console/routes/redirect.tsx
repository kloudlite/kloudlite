import { redirect } from '@remix-run/node';
import getQueries from '~/root/lib/server/helpers/get-queries';

export const loader = (ctx: any) => {
  const { url } = getQueries(ctx);
  return redirect(url);
};
