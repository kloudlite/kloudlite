const baseUrls = () => {
  const bUrl =
    (() => {
      if (typeof window !== 'undefined') {
        // @ts-ignore
        return window.BASE_URL;
      }
      return process.env.BASE_URL;
    })() || 'kloudlite.io';

  const postFix =
    (() => {
      if (typeof window !== 'undefined') {
        // @ts-ignore
        return window.URL_SUFFIX;
      }
      return process.env.URL_SUFFIX;
    })() || '';

  const cookieDomain =
    (() => {
      if (typeof window !== 'undefined') {
        // @ts-ignore
        return window.COOKIE_DOMAIN;
      }
      return process.env.COOKIE_DOMAIN;
    })() || '.kloudlite.io';

  return {
    gatewayUrl: 'http://gateway-api.kl-core.svc.cluster.local',
    authBaseUrl: `https://auth${postFix}.${bUrl}`,
    consoleBaseUrl: `https://console${postFix}.${bUrl}`,
    cookieDomain,
    baseUrl: bUrl,
    registryUrl: `registry.${bUrl}`,
    githubAppName: 'kloudlite-dev',
  };
};

const defaultConfig = {
  gatewayUrl: baseUrls().gatewayUrl,
  authBaseUrl: baseUrls().authBaseUrl,
  consoleBaseUrl: baseUrls().consoleBaseUrl,
  cookieDomain: baseUrls().cookieDomain,
  baseUrl: baseUrls().baseUrl,
  registryUrl: baseUrls().registryUrl,
  githubAppName: baseUrls().githubAppName,
};

module.exports = defaultConfig;
