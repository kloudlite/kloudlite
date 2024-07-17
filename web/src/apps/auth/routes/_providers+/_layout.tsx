import { IRemixCtx } from '~/root/lib/types/common';
import logger from '~/root/lib/client/helpers/log';
import {
  Outlet,
  ShouldRevalidateFunction,
  useLoaderData,
} from '@remix-run/react';
import { assureNotLoggedIn } from '~/root/lib/server/helpers/minimal-auth';
import { GQLServerHandler } from '~/auth/server/gql/saved-queries';

export type IProviderContext = {
  githubLoginUrl: string;
  gitlabLoginUrl: string;
  googleLoginUrl: string;
};

const Provider = () => {
  const data = useLoaderData();
  return <Outlet context={data} />;
};

export const restActions = async (ctx: IRemixCtx) => {
  const { data: checkData, errors: checkError } = await GQLServerHandler(
    ctx.request
  ).checkOauthEnabled({});

  if (checkError) {
    logger.error(checkError);
  }

  const { data, errors } = await GQLServerHandler(
    ctx.request
  ).loginPageInitUrls();

  if (errors) {
    logger.error(errors);
  }

  const {
    githubLoginUrl = '',
    gitlabLoginUrl = '',
    googleLoginUrl = '',
  } = data || {};

  return {
    githubLoginUrl: checkData?.find(
      (v) => v.provider === 'github' && !v.enabled
    )
      ? ''
      : githubLoginUrl,
    gitlabLoginUrl: checkData?.find(
      (v) => v.provider === 'gitlab' && !v.enabled
    )
      ? ''
      : gitlabLoginUrl,
    googleLoginUrl: checkData?.find(
      (v) => v.provider === 'google' && !v.enabled
    )
      ? ''
      : googleLoginUrl,
  };
};

export const loader = async (ctx: IRemixCtx) =>
  (await assureNotLoggedIn(ctx)) || restActions(ctx);

export const shouldRevalidate: ShouldRevalidateFunction = ({
  currentUrl,
  nextUrl,
  defaultShouldRevalidate,
}) => {
  if (!defaultShouldRevalidate) {
    return false;
  }
  if (currentUrl.search !== nextUrl.search) {
    return false;
  }
  return true;
};

export default Provider;
