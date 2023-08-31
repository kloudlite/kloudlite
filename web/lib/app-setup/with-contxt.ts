// import logger from '../client/helpers/log';

import { redirect } from 'react-router-dom';
import { ExtRCtxProps, MapType } from '../types/common';

const withContext = (
  ctx: ExtRCtxProps,
  props: MapType,
  headers: MapType = {}
) => {
  let _props = props;

  if (ctx.authProps) {
    // logger.log('calling');
    _props = ctx.authProps(props);
  }

  if (ctx.consoleContextProps) {
    _props = ctx.consoleContextProps(_props);
  }

  return new Response(JSON.stringify(_props), {
    headers: {
      'Content-Type': 'application/json',
      'set-cookie': [...(ctx.request.cookies || [])].join('; '),
      ...headers,
    },
  });
};

export const redirectWithContext = (
  ctx: ExtRCtxProps,
  path: string,
  headers = {}
) => {
  return redirect(path, {
    headers: {
      'Set-Cookie': [...(ctx.request.cookies || [])].join('; '),
      ...headers,
    },
  });
};

export default withContext;
