import { getMetadata } from './common';

export const getSpotSpecs = ({ memMin, memMax, cpuMin, cpuMax }) => ({
  ...{ memMin, memMax, cpuMin, cpuMax },
});

export const getOnDemandSpecs = ({ instanceType }) => ({ ...{ instanceType } });

export const getAwsNodeConfig = (
  {
    imageId,
    provisionMode,
    region,
    vpc = '',
    spotSpecs = undefined,
    onDemandSpecs = undefined,
  } = {
    provisionMode: 'on_demand' || 'reserved' || 'spot',
    region: '',
    imageId: '',
  }
) => ({
  ...{ imageId, provisionMode, region, vpc, spotSpecs, onDemandSpecs },
});

export const getNodepoolSpec = (
  { awsNodeConfig = undefined, maxCount, minCount } = {
    awsNodeConfig: getAwsNodeConfig(),
    minCount: 0,
    maxCount: 0,
  }
) => ({
  ...{ awsNodeConfig, maxCount, minCount, targetCount: minCount },
});

export const getNodepool = (
  { metadata, spec } = {
    metadata: getMetadata(),
    spec: getNodepoolSpec(),
  }
) => ({
  ...{ metadata, spec },
});
