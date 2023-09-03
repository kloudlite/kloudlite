// @ts-ignore
import { withRPC } from '@madhouselabs/madrpc';
import { IRemixCtx } from '../types/common';

export const RootAPIAction =
  (GQLServerHandler: any) => async (ctx: IRemixCtx) => {
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

    const {
      data,
      errors,
      cookie,
    }: {
      data: any;
      errors: any;
      cookie: any;
    } = await new Promise((resolve, reject) => {
      middleware(
        { body: reqData },
        {
          json: (
            _data:
              | { data: any; errors: any; cookie: any }
              | PromiseLike<{ data: any; errors: any; cookie: any }>
          ) => {
            resolve(_data);
          },
        },
        (err: any) => {
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
