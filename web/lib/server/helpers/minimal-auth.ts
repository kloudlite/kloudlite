import { redirect } from '@remix-run/node';
import { GQLServerHandler } from '../gql/saved-queries';
import { authBaseUrl, consoleBaseUrl } from '../../configs/base-url.cjs';
import { getCookie } from '../../app-setup/cookies';
import { redirectWithContext } from '../../app-setup/with-contxt';
import { IExtRemixCtx, MapType, IRemixReq } from '../../types/common';

export const assureNotLoggedIn = async (ctx: { request: IRemixReq }) => {
  const whoAmI = await GQLServerHandler({
    headers: ctx.request.headers,
  }).whoAmI();

  if (whoAmI.data) {
    return redirect(`/`);
  }

  return false;
};

export const minimalAuth = async (ctx: IExtRemixCtx) => {
  const cookie = getCookie(ctx);

  const whoAmI = await GQLServerHandler({
    headers: ctx.request?.headers,
  }).whoAmI();

  if (
    whoAmI.errors &&
    whoAmI.errors[0].message === 'input: auth_me user not logged in'
  ) {
    return redirect(`${authBaseUrl}/login`);
  }

  if (!(whoAmI.data && whoAmI.data)) {
    if (new URL(ctx.request.url).host === new URL(consoleBaseUrl).host) {
      const { pathname } = new URL(ctx.request.url);
      const history = cookie.get('url_history');
      if (history !== pathname) {
        cookie.remove('url_history');
        return redirectWithContext(ctx, pathname);
      }
    }
    // set to history so we can redirect to same page again

    return redirect(`${authBaseUrl}/login`);
  }

  if (!(whoAmI.data && whoAmI.data.verified)) {
    return redirect(`${authBaseUrl}/verify-email`);
  }

  ctx.authProps = (props: MapType) => {
    return {
      ...props,
      user: whoAmI.data,
    };
  };

  return false;
};
