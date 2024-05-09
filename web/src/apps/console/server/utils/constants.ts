import { registryHost } from '~/root/lib/configs/base-url.cjs';

export const constants = {
  nan: 'NaN',
  cacheRepoName: 'kloudlite-cache',
  kloudliteHelmAgentName: 'kloudlite-agent',
  defaultAppRepoName: (account: string) =>
    `${registryHost}/${account}/kloudlite-apps`,
  defaultAppRepoNameOnly: 'kloudlite-apps',
  metadot: '·',
};
