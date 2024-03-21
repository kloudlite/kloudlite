import { serverError } from '../../server/helpers/server-error';

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

export const isDev = getNodeEnv() === 'development';

const logger = {
  time: isDev ? console.time : () => {},
  timeEnd: isDev ? console.timeEnd : () => {},
  log: isDev ? console.log : () => {},

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
      if (!isDev) {
        console.trace(`\n\n${err}\n\n`);
        return;
      }
      console.error(`\n\n${err}\n\n`);
      return;
    }

    console.trace(`\n\n${args}\n\n`);

    if (isDev && typeof window === 'undefined') {
      serverError(args);
    }
  },
};

export default logger;
