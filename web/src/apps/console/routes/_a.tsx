import { Outlet, useLoaderData, useOutletContext } from '@remix-run/react';
import { getCookie } from '~/root/lib/app-setup/cookies';
import withContext, {
  redirectWithContext,
} from '~/root/lib/app-setup/with-contxt';
import { minimalAuth } from '~/root/lib/server/helpers/minimal-auth';
import { IExtRemixCtx } from '~/root/lib/types/common';

const Auth = () => {
  const { user } = useLoaderData();
  const rootContext: any = useOutletContext();
  return <Outlet context={{ ...rootContext, user }} />;
};

const restActions = (ctx: IExtRemixCtx) => {
  // redirect to history if available
  const cookie = getCookie(ctx);
  const history = cookie.get('url_history');
  if (history) {
    cookie.remove('url_history');
    return redirectWithContext(ctx, history);
  }

  return withContext(ctx, {});
};

export const loader = async (ctx: IExtRemixCtx) => {
  return (await minimalAuth(ctx)) || restActions(ctx);
};

export default Auth;
