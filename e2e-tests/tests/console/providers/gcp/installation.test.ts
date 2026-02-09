import { consoleTest } from '../../../../lib/fixtures'
import { runProviderInstallationTest } from '../../../../lib/provider-test'

consoleTest.describe('Console > Providers > GCP > Installation', () => {
  consoleTest.use({ storageState: { cookies: [], origins: [] } })

  runProviderInstallationTest(consoleTest, {
    name: 'GCP',
    tabName: 'GCP',
    prerequisiteText: 'gcloud CLI configured with Application Default Credentials',
    regionLabel: 'Select GCP Region:',
    installUrlPath: 'install/gcp',
    uninstallUrlPath: 'uninstall/gcp',
    subdomainPrefix: 'e2e-gcp-',
    testNamePrefix: 'E2E GCP',
    tmpDirPrefix: 'kl-e2e-gcp-',
  })
})
