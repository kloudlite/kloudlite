import { print } from 'graphql';
import { gatewayUrl } from '~/root/lib/base-url';
import logger from '../../client/helpers/log';

const parseData = (data, dataPaths) => {
  if (dataPaths.length === 0) return data;
  if (!data) return data;
  return parseData(data[dataPaths[0]], dataPaths.slice(1));
};

// eslint-disable-next-line max-len
export const ExecuteQueryWithContext =
  (headers) =>
  (q, { dataPath = '', transformer = (val) => val } = {}, def = null) =>
  async (variables) => {
    try {
      const resp = await fetch(gatewayUrl, {
        method: 'POST', // or 'PUT'
        headers: {
          ...{
            cookie: headers.get('klsession') || headers.get('cookie') || null,
          },
        },
        body: {
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
        logger.error(err);
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
