import { IMiddlewareResponse, withRPC } from '../server/helpers/rpc';
import { IGQLServerProps, IRemixCtx } from '../types/common';

export const RootAPIAction =
  (
    GQLServerHandler: (props: IGQLServerProps) => {
      [x: string]: (props: any) => Promise<any>;
    }
  ) =>
  async (ctx: IRemixCtx) => {
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

    const { data, errors, cookie }: IMiddlewareResponse = await new Promise(
      (resolve, reject) => {
        middleware(
          { body: reqData },
          {
            json: (_data) => {
              resolve(_data);
            },
          },
          (err: Error) => {
            reject(err);
          }
        );
      }
    );

    return new Response(JSON.stringify({ data, errors }), {
      headers: {
        'Content-Type': 'application/json',
        ...(cookie ? { 'set-cookie': cookie } : {}),
      },
    });
  };
