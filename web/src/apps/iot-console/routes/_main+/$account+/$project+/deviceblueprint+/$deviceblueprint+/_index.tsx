import { redirect } from '@remix-run/node';
import { IRemixCtx } from '~/root/lib/types/common';

export const loader = async (ctx: IRemixCtx) => {
  const { account,  deviceblueprint } = ctx.params;
  return redirect(
    `/${account}/deviceblueprint/${deviceblueprint}/apps`
  );
};
