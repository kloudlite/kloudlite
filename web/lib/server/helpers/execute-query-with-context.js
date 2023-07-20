import { print } from 'graphql';
import axios from 'axios';
import { gatewayUrl } from '../../configs/base-url.cjs';

const parseData = (data, dataPaths) => {
  if (dataPaths.length === 0) return data;
  if (!data) return data;
  return parseData(data[dataPaths[0]], dataPaths.slice(1));
};

export const ExecuteQueryWithContext =
  (headers) =>
  (q, { dataPath = '', transformer = (val) => val } = {}, def = null) =>
  async (variables) => {
    try {
      const resp = await axios({
        url: gatewayUrl,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json; charset=utf-8',
          ...{
            cookie: headers.get('klsession') || headers.get('cookie') || null,
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
