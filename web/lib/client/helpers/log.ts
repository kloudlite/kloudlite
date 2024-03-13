// import axios from 'axios';
// import { consoleBaseUrl } from '../../configs/base-url.cjs';
import { serverError } from '../../server/helpers/server-error';
// import { parseError } from '../../utils/common';

const getNodeEnv = () => {
  const env = (() => {
    if (typeof window !== 'undefined') {
      // @ts-ignore
      return window.NODE_ENV;
    }
    return process.env.NODE_ENV;
  })();

  if (env) {
    return env;
  }

  return 'development';
};

/* eslint-disable no-unused-vars */
/* eslint-disable @typescript-eslint/no-unused-vars */
export const PostErr = async (message: string, source: string) => {
  // try {
  //   await axios.post(
  //     'https://hooks.slack.com/services/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',
  //     {
  //       body: {
  //         channel: source === 'server' ? '#bugs' : '#web-errors',
  //         username:
  //           typeof window === 'undefined' ? 'server-error' : 'web-error',
  //         text: message,
  //         icon_emoji: ':ghost:',
  //       },
  //     }
  //   );
  // } catch (err) {
  //   console.log(parseError(err).message);
  // }
  return {};
};

const PostToHook = (message: string) => {
  // if (typeof window === 'undefined') {
  //   return PostErr(message, 'server');
  // }
  //
  // try {
  //   axios.post(`${consoleBaseUrl}/api/error`, {
  //     body: { error: message },
  //   });
  // } catch (err) {
  //   console.log(err);
  // }
  return {};
};

export const isDev = getNodeEnv() === 'development';

const logger = {
  time: isDev ? console.time : () => {},
  timeEnd: isDev ? console.timeEnd : () => {},
  log: isDev ? console.log : () => {},

  // log: console.log,

  warn: console.warn,
  trace: (...args: any[]) => {
    let err;
    try {
      err = JSON.stringify(args, null, 2);
    } catch (_) {
      console.log('');
    }

    if (err) {
      console.trace(err);
    } else {
      console.trace(args);
    }
  },

  error: (...args: any[]) => {
    let err;
    try {
      err = JSON.stringify(args, null, 2);
    } catch (_) {
      console.log('');
    }

    if (err) {
      console.trace(`\n\n${err}\n\n`);
      if (!isDev) {
        PostToHook(`\`\`\`${err}\`\`\``);
      }
    } else {
      console.trace(`\n\n${args}\n\n`);
    }

    if (isDev && typeof window === 'undefined') {
      serverError(args);
    }
  },
};

export default logger;
