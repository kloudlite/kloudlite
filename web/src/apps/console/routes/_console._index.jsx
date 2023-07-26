import { redirect } from '@remix-run/node';
import { setupConsoleContext } from '../server/utils/auth-utils';

const restActions = async (ctx) => {
  return redirect('/projects');
};

export const loader = async (ctx) => {
  return (await setupConsoleContext(ctx)) || restActions(ctx);
};
