import { consoleTest } from '../../../../lib/fixtures'
import { runProviderInstallationTest } from '../../../../lib/provider-test'

consoleTest.describe('Console > Providers > Azure > Installation', () => {
  consoleTest.use({ storageState: { cookies: [], origins: [] } })

  runProviderInstallationTest(consoleTest, {
    name: 'Azure',
    tabName: 'Azure',
    prerequisiteText: 'Azure CLI configured',
    regionLabel: 'Select Azure Location:',
    installUrlPath: 'install/azure',
    uninstallUrlPath: 'uninstall/azure',
    subdomainPrefix: 'e2e-az-',
    testNamePrefix: 'E2E Azure',
    tmpDirPrefix: 'kl-e2e-az-',
  })
})
