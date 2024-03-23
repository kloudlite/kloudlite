// @ts-nocheck
const getClientEnv = (env) => {
  const {
    BASE_URL,
    NODE_ENV,
    URL_SUFFIX,
    DEVELOPER,
    REGISTRY_URL,
    MANAGE_GITLAB_URL,
    MANAGE_GITHUB_URL,
    GITHUB_APP_NAME,
  } = env;
  return `
${BASE_URL ? `window.BASE_URL = ${`'${BASE_URL}'`}` : ''}
${
  NODE_ENV === 'development'
    ? `window.DEVELOPER = ${`'${DEVELOPER}'`}`
    : `window.NODE_ENV = ${`'${NODE_ENV}'`}`
}
${URL_SUFFIX ? `window.URL_SUFFIX = ${`'${URL_SUFFIX}'`}` : ''}
${REGISTRY_URL ? `window.REGISTRY_URL = ${`'${REGISTRY_URL}'`}` : ''}
${
  MANAGE_GITHUB_URL
    ? `window.MANAGE_GITHUB_URL = ${`'${MANAGE_GITHUB_URL}'`}`
    : ''
}
${
  MANAGE_GITLAB_URL
    ? `window.MANAGE_GITLAB_URL = ${`'${MANAGE_GITLAB_URL}'`}`
    : ''
}
${GITHUB_APP_NAME ? `window.GITHUB_APP_NAME = ${`'${GITHUB_APP_NAME}'`}` : ''}
`;
};

const getServerEnv = () => {
  const nodeEnv = process.env.NODE_ENV;
  return {
    NODE_ENV: nodeEnv,
    ...(nodeEnv === 'development'
      ? { PORT: Number(process.env.PORT), DEVELOPER: process.env.DEVELOPER }
      : {}),

    ...(process.env.URL_SUFFIX ? { URL_SUFFIX: process.env.URL_SUFFIX } : {}),
    ...(process.env.BASE_URL ? { BASE_URL: process.env.BASE_URL } : {}),
    ...(process.env.GATEWAY_URL
      ? { GATEWAY_URL: process.env.GATEWAY_URL }
      : {}),
    ...(process.env.REGISTRY_URL
      ? { REGISTRY_URL: process.env.REGISTRY_URL }
      : {}),
    ...(process.env.MANAGE_GITHUB_URL
      ? { MANAGE_GITHUB_URL: process.env.MANAGE_GITHUB_URL }
      : {}),
    ...(process.env.MANAGE_GITLAB_URL
      ? { MANAGE_GITLAB_URL: process.env.MANAGE_GITLAB_URL }
      : {}),
    ...(process.env.GITHUB_APP_NAME
      ? { GITHUB_APP_NAME: process.env.GITHUB_APP_NAME }
      : {}),
  };
};

const baseUrls = () => {
  const bUrl =
    (() => {
      if (typeof window !== 'undefined') {
        return window.BASE_URL;
      }
      return process.env.BASE_URL;
    })() || 'kloudlite.io';

  const postFix =
    (() => {
      if (typeof window !== 'undefined') {
        return window.URL_SUFFIX;
      }
      return process.env.URL_SUFFIX;
    })() || '';

  const cookieDomain =
    (() => {
      if (typeof window !== 'undefined') {
        return window.COOKIE_DOMAIN;
      }
      return process.env.COOKIE_DOMAIN;
    })() || '.kloudlite.io';

  const gatewayUrl =
    (() => {
      if (typeof window !== 'undefined') {
        return window.GATEWAY_URL;
      }
      return process.env.GATEWAY_URL;
    })() || 'http://gateway-api.kl-core.svc.cluster.local';

  const registryHost =
    (() => {
      if (typeof window !== 'undefined') {
        return window.REGISTRY_URL;
      }
      return process.env.REGISTRY_URL;
    })() || `registry.${bUrl}`;

  const gitEnvs = (() => {
    if (typeof window !== 'undefined') {
      return {
        githubAppName: window.GITHUB_APP_NAME,
      };
    }
    return {
      githubAppName: process.env.GITHUB_APP_NAME,
    };
  })() || {
    githubManageUrl: '',
    gitlabManageUrl: '',
    githubAppName: '',
  };

  return {
    gatewayUrl,
    authBaseUrl: `https://auth${postFix}.${bUrl}`,
    consoleBaseUrl: `https://console${postFix}.${bUrl}`,
    registryHost,
    cookieDomain,
    baseUrl: bUrl,
    socketUrl: `wss://websocket.${bUrl}`,
    observeUrl: `https://observe.${bUrl}`,
    gitEnvs,
  };
};

const defaultConfig = {
  gatewayUrl: baseUrls().gatewayUrl,
  authBaseUrl: baseUrls().authBaseUrl,
  consoleBaseUrl: baseUrls().consoleBaseUrl,
  cookieDomain: baseUrls().cookieDomain,
  baseUrl: baseUrls().baseUrl,
  socketUrl: baseUrls().socketUrl,
  registryHost: baseUrls().registryHost,
  observeUrl: baseUrls().observeUrl,
  getServerEnv,
  getClientEnv,
  gitEnvs: baseUrls().gitEnvs,
};

module.exports = defaultConfig;
