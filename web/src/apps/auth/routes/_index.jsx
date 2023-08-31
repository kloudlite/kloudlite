import logger from '~/root/lib/client/helpers/log';
import { redirect } from '@remix-run/node';
import { authBaseUrl, consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { minimalAuth } from '~/root/lib/server/helpers/minimal-auth';
import { getCookie } from '~/root/lib/app-setup/cookies';
import { GQLServerHandler } from '../server/gql/saved-queries';

// @ts-ignore
const restActions = async (ctx) => {
  const cookie = getCookie(ctx);
  if (cookie.get('cliLogin')) {
    try {
      const { data, errors } = await GQLServerHandler(
        ctx.request
      ).setRemoteAuthHeader({
        loginId: cookie.get('cliLogin'),
        authHeader: ctx?.request?.headers?.get('cookie'),
      });
      logger.log(data, 'loggedin');
      if (errors) {
        throw errors[0];
      }

      return redirect(`${authBaseUrl}/cli-logged-in`);
    } catch (err) {
      logger.error(err);
      return {
        err,
      };
    }
  }
  return redirect(consoleBaseUrl);
};

export const loader = async (/** @type {any} */ ctx) => {
  return (await minimalAuth(ctx)) || restActions(ctx);
};
