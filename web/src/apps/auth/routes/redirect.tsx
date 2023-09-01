import { redirect } from '@remix-run/node';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { IRemixCtx } from '~/root/lib/types/common';

export const loader = (ctx: IRemixCtx) => {
  const { url } = getQueries(ctx);
  return redirect(url);
};
