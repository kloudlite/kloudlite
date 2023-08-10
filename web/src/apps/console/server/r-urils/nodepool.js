import { getMetadata } from './common';

export const getSpotSpecs = (
  { memMin, memMax, cpuMin, cpuMax } = {
    memMin: 0,
    memMax: 0,
    cpuMin: 0,
    cpuMax: 0,
  }
) => ({
  ...{ memMin, memMax, cpuMin, cpuMax },
});

export const getOnDemandSpecs = (
  { instanceType } = {
    instanceType: 'c6a-large',
  }
) => ({
  ...{ instanceType },
});

export const getAwsNodeConfig = (
  {
    provisionMode,
    region,
    vpc = '',
    spotSpecs = null,
    onDemandSpecs = null,
  } = {
    provisionMode: 'on_demand' || 'reserved' || 'spot',
    region: '',
    spotSpecs: getSpotSpecs(),
    onDemandSpecs: getOnDemandSpecs(),
  }
) => ({
  ...{ provisionMode, region, vpc, spotSpecs, onDemandSpecs },
});

export const getNodePoolSpec = (
  { awsNodeConfig = undefined, maxCount, minCount } = {
    awsNodeConfig: getAwsNodeConfig(),
    minCount: 0,
    maxCount: 0,
  }
) => ({
  ...{ awsNodeConfig, maxCount, minCount, targetCount: minCount },
});

export const getNodePool = (
  { metadata, spec } = {
    metadata: getMetadata(),
    spec: getNodePoolSpec(),
  }
) => ({
  ...{ metadata, spec },
});

// parsing things
export const parseProvisionMode = (item) => {
  return item?.spec?.awsNodeConfig?.provisionMode || '';
};
export const parseCapacity = (item) => {
  return `${item?.spec?.minCount || 0} min - ${item?.spec?.maxCount || 0} max`;
};
