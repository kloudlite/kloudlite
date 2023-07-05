import { redirect } from '@remix-run/node';
import { authBaseUrl } from '~/root/lib/base-url';
import logger from '../../client/helpers/log';
import { GQLServerHandler } from '../gql/saved-queries';

export const assureNotLoggedIn = async ({ ctx }) => {
  const rand = `${Math.random()}`;
  logger.time(`${rand}:whoami`);
  const whoAmI = await GQLServerHandler({
    headers: ctx?.request?.headers,
  }).whoAmI();

  logger.timeEnd(`${rand}:whoami`);

  if (whoAmI.data && whoAmI.data.me) {
    return redirect(`/`);
  }
  return false;
};

export const minimalAuth = async ({ ctx }) => {
  const rand = `${Math.random()}`;
  logger.time(`${rand}:whoami`);

  const whoAmI = await GQLServerHandler({
    headers: ctx?.request?.headers,
  }).whoAmI();

  logger.timeEnd(`${rand}:whoami`);


  if (!(whoAmI.data && whoAmI.data.me)) {
    return redirect(`${authBaseUrl}/login`);
  }

  if (!(whoAmI.data && whoAmI.data.me.verified)) {
    return redirect(`${authBaseUrl}/verify-email`);
  }

  ctx.authProps = (props) => {
    return {
      ...props,
      user: whoAmI.data.me,
    };
  };
  return false;
};
