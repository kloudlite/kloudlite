import { Outlet } from '@remix-run/react';
import { getCookie } from '~/root/lib/app-setup/cookies';
import withContext, {
  redirectWithContext,
} from '~/root/lib/app-setup/with-contxt';
import { minimalAuth } from '~/root/lib/server/helpers/minimal-auth';

const Auth = () => {
  return <Outlet />;
};

const restActions = (ctx) => {
  // redirect to history if available
  const cookie = getCookie(ctx);
  const history = cookie.get('url_history');
  if (history) {
    cookie.remove('url_history');
    return redirectWithContext(ctx, history);
  }
  return withContext(ctx, {});
};

export const loader = async (ctx) => {
  return (await minimalAuth(ctx)) || restActions(ctx);
};

export default Auth;
