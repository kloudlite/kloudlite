import axios from 'axios';
import { MapType } from '~/root/lib/types/common';

export interface IMiddlewareResponse {
  data: MapType<string | number | boolean>;
  errors: Error[];
  cookie: any;
}

export const withRPC = (
  handler: {
    [key: string]: (...args: any) => Promise<any>;
  },
  _options?: MapType<any>
) => {
  return async (
    req: {
      body: {
        method?: string;
        args?: MapType<string | number | boolean>[];
      };
    },
    res: {
      json: (
        arg0: IMiddlewareResponse | PromiseLike<IMiddlewareResponse>
      ) => void;
    },
    next: (arg0: Error) => void
  ) => {
    try {
      if (!req.body.method) throw new Error('Handler Method not found');

      const method = req.body.method.split('.').reduce((acc: any, item) => {
        return acc[item];
      }, handler);

      const response = await method(...(req.body.args || []));
      res.json(response);
    } catch (err) {
      next(err as Error);
    }
  };
};

const makeCall = async ({
  url,
  method,
  args,
  httpOptions,
}: {
  url: string;
  method?: string;
  args: any;
  httpOptions: any;
}) => {
  const { data } = await axios.post(
    url,
    {
      method,
      args,
    },
    httpOptions
  );
  return data;
};

export const createRPCClient = (
  url: string,
  slack?: string,
  options?: any
): any => {
  return new Proxy(() => {}, {
    apply(target, method, args) {
      return makeCall({
        url,
        method: slack,
        args,
        httpOptions: options,
      });
    },
    get(target, prefix) {
      return createRPCClient(
        url,
        slack ? [slack, prefix].join('.') : (prefix as string),
        options
      );
    },
  }) as any;
};
