import { withRPC } from '@madhouselabs/madrpc';

export const RootAPIAction = (GQLServerHandler) => async (ctx) => {
  if (ctx.request.method !== 'POST') {
    return new Response(JSON.stringify({ message: 'Not Found' }), {
      status: 404,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  const middleware = withRPC(
    GQLServerHandler({ headers: ctx.request.headers })
  );

  const reqData = await ctx.request.json();

  const { data, errors, cookie } = await new Promise((resolve, reject) => {
    middleware(
      { body: reqData },
      {
        json: (_data) => {
          resolve(_data);
        },
      },
      (err) => {
        reject(err);
      }
    );
  });

  return new Response(JSON.stringify({ data, errors }), {
    headers: {
      'Content-Type': 'application/json',
      ...(cookie ? { 'set-cookie': cookie } : {}),
    },
  });
};
