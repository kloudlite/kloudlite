const baseUrls = () => {
  const defaultBaseUrl = 'dev.kloudlite.io';

  // const bUrl = (() => {
  //   if (typeof window !== 'undefined') {
  //     // @ts-ignore
  //     if (window.BASE_URL === 'undefined') {
  //       return defaultBaseUrl;
  //     }
  //     // @ts-ignore
  //     return window.BASE_URL || defaultBaseUrl;
  //   }
  //   return process.env.BASE_URL || defaultBaseUrl;
  // })();

  const bUrl = defaultBaseUrl;

  return {
    gatewayUrl: 'http://gateway-api.kl-core.svc.cluster.local',
    authBaseUrl: `https://auth.${bUrl}`,
    consoleBaseUrl: `https://console.${bUrl}`,
    cookieDomain: `.kloudlite.io`,
    baseUrl: bUrl,
  };
};

const defaultConfig = {
  gatewayUrl: baseUrls().gatewayUrl,
  authBaseUrl: baseUrls().authBaseUrl,
  consoleBaseUrl: baseUrls().consoleBaseUrl,
  cookieDomain: baseUrls().cookieDomain,
  baseUrl: baseUrls().baseUrl,
};

module.exports = defaultConfig;
