import { redirect } from '@remix-run/node';
import logger from '../../client/helpers/log';
import { GQLServerHandler } from '../gql/saved-queries';
import { authBaseUrl, consoleBaseUrl } from '../../configs/base-url.cjs';
import { getCookie } from '../../app-setup/cookies';
import { redirectWithContext } from '../../app-setup/with-contxt';
import { IExtRemixCtx, MapType, IRemixReq } from '../../types/common';

export const assureNotLoggedIn = async (ctx: { request: IRemixReq }) => {
  const rand = `${Math.random()}`;
  logger.time(`${rand}:whoami`);
  const whoAmI = await GQLServerHandler({
    headers: ctx.request.headers,
  }).whoAmI();

  logger.timeEnd(`${rand}:whoami`);

  if (whoAmI.data && whoAmI.data.me) {
    return redirect(`/`);
  }
  return false;
};

export const minimalAuth = async (ctx: IExtRemixCtx) => {
  const rand = `${Math.random()}`;
  logger.time(`${rand}:whoami`);
  const cookie = getCookie(ctx);

  const whoAmI = await GQLServerHandler({
    headers: ctx.request?.headers,
  }).whoAmI();

  if (whoAmI.errors && whoAmI.errors[0].message === 'user not logged in') {
    return redirect(`${authBaseUrl}/login`);
  }

  logger.timeEnd(`${rand}:whoami`);

  if (!(whoAmI.data && whoAmI.data.me)) {
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

  if (!(whoAmI.data && whoAmI.data.me.verified)) {
    return redirect(`${authBaseUrl}/verify-email`);
  }

  ctx.authProps = (props: MapType) => {
    return {
      ...props,
      user: whoAmI.data.me,
    };
  };

  return false;
};
