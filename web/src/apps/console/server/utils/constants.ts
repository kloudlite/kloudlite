import { registryHost } from '~/root/lib/configs/base-url.cjs';

export const constants = {
  nan: 'NaN',
  cacheRepoName: 'kloudlite-cache',
  kloudliteHelmAgentName: 'kloudlite-agent',
  defaultAppRepoName: (account: string) =>
    `${registryHost}/${account}/kloudlite-apps`,
  defaultAppRepoNameOnly: 'kloudlite-apps',
  metadot: '·',
  dockerImageFormatRegex:
    /^(([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*\.)+[a-zA-Z]{2,}(:[0-9]+)?\/)?([a-z0-9]+(-[a-z0-9]+)*\/)*[a-z0-9]+([._-][a-z0-9]+)*(:[a-zA-Z0-9_][a-zA-Z0-9._-]{0,127}[a-zA-Z0-9])?(@[A-Za-z][A-Za-z0-9]*(?:[._-][A-Za-z0-9]+)?:[A-Fa-f0-9]{32,})?$/,
  keyFormatRegex: /^[A-Za-z0-9_]+([./-][A-Za-z0-9_]+)*$/,
  kloudliteClusterName: 'kloudlite-enabled-cluster',
};
