import logger from '~/root/lib/client/helpers/log';
import { redirect } from '@remix-run/node';
import { authBaseUrl, consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { minimalAuth } from '~/root/lib/server/helpers/minimal-auth';
import { getCookie } from '~/root/lib/app-setup/cookies';
import { IExtRemixCtx } from '~/root/lib/types/common';
import { GQLServerHandler } from '../server/gql/saved-queries';

const restActions = async (ctx: IExtRemixCtx) => {
  const cookie = getCookie(ctx);
  if (cookie.get('cliLogin')) {
    try {
      console.log('here', 'lskdfjsldfj');

      const { data, errors } = await GQLServerHandler(
        ctx.request
      ).setRemoteAuthHeader({
        loginId: cookie.get('cliLogin') || '',
        authHeader: ctx?.request?.headers?.get('cookie'),
      });

      logger.log(data, 'loggedin');
      if (errors) {
        console.log('here', errors);
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

export const loader = async (ctx: IExtRemixCtx) => {
  return (await minimalAuth(ctx)) || restActions(ctx);
};
