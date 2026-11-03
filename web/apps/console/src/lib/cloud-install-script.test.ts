import { describe, expect, it } from 'vitest'
import { getCloudInstallScript } from './cloud-install-script'

describe('getCloudInstallScript', () => {
  it('builds provider-specific install and uninstall commands', () => {
    expect(getCloudInstallScript('aws', 'install')).toContain(
      'kli install aws ${REGION:+--region "$REGION"}',
    )
    expect(getCloudInstallScript('azure', 'uninstall')).toContain(
      'kli uninstall azure ${LOC:+--location "$LOC"}',
    )
    expect(getCloudInstallScript('gcp', 'install')).toContain('${ZONE:+--zone "$ZONE"}')
  })

  it('rejects unknown providers', () => {
    expect(getCloudInstallScript('unknown', 'install')).toBeNull()
  })
})
