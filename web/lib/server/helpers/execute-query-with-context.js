import { print } from 'graphql';
import ServerCookie from 'cookie';
import axios from 'axios';
import { gatewayUrl } from '../../configs/base-url.cjs';

const parseData = (data, dataPaths) => {
  if (dataPaths.length === 0) return data;
  if (!data) return data;
  return parseData(data[dataPaths[0]], dataPaths.slice(1));
};

const parseCookie = (cookieString) => {
  const [cookie] = cookieString.split(';');
  const [name, value] = cookie.split('=');
  return { name, value };
};

export const ExecuteQueryWithContext =
  (headers, cookies = []) =>
  (q, { dataPath = '', transformer = (val) => val } = {}, def = null) =>
  async (variables) => {
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
          variables,
        },
      });

      let { data } = resp.data;

      if (dataPath) {
        data = parseData(
          data,
          dataPath.split('.').filter((item) => item)
        );
      }

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
      if (err.response) {
        return err.response.data;
      }

      return {
        errors: [
          {
            message: err.message,
          },
        ],
      };
    }
  };
