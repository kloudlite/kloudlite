import { ASTNode, print } from 'graphql';
import ServerCookie from 'cookie';
import axios, { AxiosError } from 'axios';
import { gatewayUrl } from '../../configs/base-url.cjs';
import {
  ICookies,
  MapType,
  IRemixHeader,
  IGqlReturn,
  NN,
} from '../../types/common';

const parseCookie = (cookieString: string) => {
  const [cookie] = cookieString.split(';');
  const [name, value] = cookie.split('=');
  return { name, value };
};

export type IExecutorResp<B = any, C = MapType<any>> = (
  variables?: C
) => Promise<IGqlReturn<NN<B>>>;

type formatter<A, B, C> = {
  transformer: (data: A) => B;
  vars?: (_: C) => void;
  k?: string;
};

export type IExecutor = <A, B, C = MapType<any>>(
  q: ASTNode,
  formatter: formatter<A, B, C>,
  def?: any
) => IExecutorResp<B, C>;

export const ExecuteQueryWithContext = (
  headers: IRemixHeader,
  cookies: ICookies = []
) => {
  return function executor<A, B, C = MapType<any>>(
    q: ASTNode,
    formatter: formatter<A, B, C>,
    def?: any
  ): IExecutorResp<B, C> {
    const apiName = `[API] ${
      (q as any)?.definitions[0]?.selectionSet?.selections[0]?.name?.value || ''
    }`;

    const res: IExecutorResp<B, C> = async (variables) => {
      const { transformer } = formatter;

      console.time(apiName);

      try {
        const defCookie =
          headers.get('klsession') || headers.get('cookie') || null;

        const cookie = ServerCookie.parse(defCookie || '');

        if (cookies.length > 0) {
          for (let i = 0; i < cookies.length; i += 1) {
            const { name, value } = parseCookie(cookies[i]);
            cookie[name] = value;
          }
        }

        const resp = await axios({
          url: gatewayUrl,
          method: 'POST',
          headers: {
            'Content-Type': 'application/json; charset=utf-8',
            ...{
              cookie: Object.entries(cookie)
                .map(([key, value]) => `${key}=${value}`)
                .join('; '),
            },
          },
          data: {
            query: print(q),
            variables: variables || {},
          },
        });

        let { data } = resp.data;

        if (data) {
          data = transformer(data);
        } else if (def) {
          data = def;
        }

        if (resp.headers && resp.headers['set-cookie']) {
          return { ...resp.data, data, cookie: resp.headers['set-cookie'] };
        }
        return { ...resp.data, data };
      } catch (err) {
        if ((err as AxiosError).response) {
          return (err as AxiosError).response?.data;
        }

        return {
          errors: [
            {
              message: (err as Error).message,
            },
          ],
        };
      } finally {
        console.timeEnd(apiName);
      }
    };

    // @ts-ignore
    res.astNode = q;
    return res;
  };
};
