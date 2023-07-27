import { Outlet } from '@remix-run/react';
import withContext from '~/root/lib/app-setup/with-contxt';
import { minimalAuth } from '~/root/lib/server/helpers/minimal-auth';

const Auth = () => {
  return <Outlet />;
};

const restActions = (ctx) => {
  return withContext(ctx, {});
};

export const loader = async (ctx) => {
  return minimalAuth(ctx) || restActions(ctx);
};

export default Auth;
