// import logger from '../client/helpers/log';

const withContext = (ctx, props, headers) => {
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
      'Set-Cookie': ctx.request.cookies || [],
      ...headers,
    },
  });
};

export default withContext;
