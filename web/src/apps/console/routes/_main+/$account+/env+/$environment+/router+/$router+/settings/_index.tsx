import { redirect } from '@remix-run/node';
import { IRemixCtx } from '~/lib/types/common';

export const loader = async (ctx: IRemixCtx) => {
  return redirect(`general`);
};
