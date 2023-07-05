import axios from 'axios';
import { consoleBaseUrl } from '../../base-url';
import { serverError } from '../../server/helpers/server-error';

const getNodeEnv = () => {
  const env = (() => {
    if (typeof window !== 'undefined') {
      // @ts-ignore
      return window.ENV;
    }
    return process.env.ENV;
  })();

  if (env) {
    return env;
  }

  return 'development';
};

export const PostErr = async (message, source) => {
  try {
    await axios({
      method: 'POST',
      url: 'https://hooks.slack.com/services/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',
      data: {
        channel: source === 'server' ? '#bugs' : '#web-errors',
        username: typeof window === 'undefined' ? 'server-error' : 'web-error',
        text: message,
        icon_emoji: ':ghost:',
      },
    });
  } catch (err) {
    console.log(err.message);
  }
  return {};
};

const PostToHook = (message) => {
  if (typeof window === 'undefined') {
    return PostErr(message, 'server');
  }

  try {
    axios({
      method: 'POST',
      url: `${consoleBaseUrl}/api/error`,
      data: { error: message },
    });
  } catch (err) {
    console.log(err);
  }
  return {};
};

const isDev = getNodeEnv() === 'development';

const logger = {
  time: isDev ? console.time : () => {},
  timeEnd: isDev ? console.timeEnd : () => {},
  log: isDev ? console.log : () => {},

  // log: console.log,

  warn: console.warn,
  trace: (...args) => {
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

  error: (...args) => {
    let err;
    try {
      err = JSON.stringify(args, null, 2);
      // axios({
      //   method: 'POST',
      //   url: 'https://hooks.slack.com/services/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',
      //   data: {
      //     channel: '#bugs',
      //     username: 'Web-Bug',
      //     text: err,
      //     icon_emoji: ':ghost:',
      //   },
      // });
    } catch (_) {
      console.log('');
    }

    if (err && isDev) {
      console.trace(err);
      PostToHook(`\`\`\`${err}\`\`\``);
    } else {
      console.trace(args);
    }

    if (isDev && typeof window === 'undefined') {
      serverError(...args);
    }
  },
};

export default logger;
