import { minimalAuth } from '~/root/lib/server/helpers/minimal-auth';
import { redirect } from 'react-router-dom';
// import logger from '~/root/lib/client/helpers/log';
import { getCookie } from '~/root/lib/app-setup/cookies';
import { GQLServerHandler } from '../gql/saved-queries';

const setTocontext = (ctx, data) => {
  ctx.consoleContextProps = (props) => ({
    ...props,
    ...data,
  });
};

const restActions = async (ctx) => {
  const ctxData = {};

  const returnData = await (async () => {
    const { account } = ctx.params;
    if (!account) {
      return redirect('/teams');
    }

    const { data: accounts, errors: asError } = await GQLServerHandler(
      ctx.request
    ).listAccounts({});

    if (asError) {
      return redirect('/teams');
    }

    ctxData.accounts = accounts;

    return false;
  })();

  setTocontext(ctx, ctxData);
  return returnData;
};

export const setupAccountContext = async (ctx) => {
  return (await minimalAuth(ctx)) || restActions(ctx);
};

export const ensureAccountSet = (ctx) => {
  const { account, a } = ctx.params;
  const cookie = getCookie(ctx);
  const cookieKey = 'kloudlite-account';

  const currentAccount = cookie.get(cookieKey);
  if (!currentAccount || currentAccount !== account) {
    cookie.set(cookieKey, account || a || '', {
      secure: true,
    });
  }
  return false;
};

export const ensureAccountClientSide = (params) => {
  const { account, a } = params;
  const cookie = getCookie();
  const cookieKey = 'kloudlite-account';

  const currentAccount = cookie.get(cookieKey);
  if (!currentAccount || currentAccount !== account) {
    cookie.set(cookieKey, account || a || '', {
      secure: true,
    });
  }
  return false;
};

export const ensureClusterSet = (ctx) => {
  const { cluster } = ctx.params;
  const cookie = getCookie(ctx);
  const cookieKey = 'kloudlite-cluster';

  const currentCluster = cookie.get(cookieKey);
  if (!currentCluster || currentCluster !== cluster) {
    cookie.set(cookieKey, cluster || '', {
      secure: true,
    });
  }
  return false;
};

export const ensureClusterClientSide = (params) => {
  const { cluster } = params;
  const cookie = getCookie();
  const cookieKey = 'kloudlite-cluster';

  const currentCluster = cookie.get(cookieKey);
  if (!currentCluster || currentCluster !== cluster) {
    cookie.set(cookieKey, cluster || '', {
      secure: true,
    });
  }
  return false;
};
