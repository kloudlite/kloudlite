import { registryHost } from '~/root/lib/configs/base-url.cjs';

export const constants = {
  nan: 'NaN',
  defaultAppRepoName: (account: string) =>
    `${registryHost}/${account}/kloudlite-apps`,
  defaultAppRepoNameOnly: 'kloudlite-apps',
};
