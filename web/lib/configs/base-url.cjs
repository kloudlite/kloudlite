const devMap = {
  bikashojha: '1',
  vision: '',
};
const baseUrls = () => {
  const bUrl = 'dev.kloudlite.io';

  const postFix = (() => {
    if (typeof window !== 'undefined') {
      return devMap[window.DEVELOPER];
    }
    return devMap[process.env.DEVELOPER];
  })();

  return {
    gatewayUrl: 'http://gateway-api.kl-core.svc.cluster.local',
    authBaseUrl: `https://auth${postFix}.${bUrl}`,
    consoleBaseUrl: `https://console${postFix}.${bUrl}`,
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
