import ServerCookie from 'cookie';
import ClientCookie from 'js-cookie';
import { cookieDomain } from '../configs/base-url.cjs';
import { IRemixCtx, MapType } from '../types/common';

export const getCookie = (ctx?: IRemixCtx) => {
  // getting all cookies
  const getAll = () => {
    if (ctx?.request) {
      return ServerCookie.parse(ctx?.request?.headers?.get('cookie') || '');
    }
    return ClientCookie.get();
  };

  // getting a cookie
  const get = (name: string) => {
    if (ctx?.request) {
      return getAll()[name];
    }
    return ClientCookie.get(name);
  };

  // setting a cookie
  const set = (name: string, value: string, options: MapType = {}) => {
    if (ctx?.request) {
      // ctx?.request.?Cookies
      if (!ctx?.request?.cookies) {
        ctx.request.cookies = [];
      }

      ctx.request.cookies.push(
        ServerCookie.serialize(name, value, {
          domain: cookieDomain,
          path: '/',
          // maxAge: 60 * 60 * 24 * 30, // 1 month
          ...options,
        })
      );

      return;
    }
    ClientCookie.set(name, value, {
      domain: cookieDomain,
      path: '/',
      ...options,
    });
  };

  // deleting a cookie
  const remove = async (name: string, options: MapType = {}) => {
    if (ctx?.request) {
      if (!ctx?.request?.cookies) {
        ctx.request.cookies = [];
      }
      ctx.request.cookies.push(
        ServerCookie.serialize(name, '--<no-value>--', {
          domain: cookieDomain,
          path: '/',
          ...options,
          maxAge: 0,
        })
      );
    }
    return ClientCookie.set(name, '', {
      domain: cookieDomain,
      path: '/',
      ...options,
      expires: 0,
    });
    // return ClientCookie.remove(name, options);
  };

  return { get, set, remove, getAll };
};
