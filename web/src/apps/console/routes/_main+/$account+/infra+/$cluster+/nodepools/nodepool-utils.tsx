import { parseValue } from '~/console/page-components/util';

const { nodePlans, provisionTypes, taintsData, gcpPoolTypes } = {
  nodePlans: [
    {
      label: 'CPU Optimised',
      options: [
        {
          label: '1x - small - 2VCPU 3.75GB Memory',
          labelDetail: {
            size: 'small',
            cpu: '2VCPU',
            memory: '4GB',
          },
          value: 'c6a.large',
          spotSpec: {
            cpuMax: 2,
            cpuMin: 2,
            memMax: 2,
            memMin: 2,
            disabled: false,
          },
          gpuEnabled: false,
        },
        {
          label: '2x - medium - 4VCPU 7.5GB Memory',
          labelDetail: {
            size: 'medium',
            cpu: '4VCPU',
            memory: '8GB',
          },
          value: 'c6a.xlarge',
          spotSpec: {
            cpuMax: 4,
            cpuMin: 4,
            memMax: 2,
            memMin: 2,
            disabled: false,
          },
        },
        {
          label: '4x - large - 8VCPU 15GB Memory',
          labelDetail: {
            size: 'large',
            cpu: '8VCPU',
            memory: '16GB',
          },
          value: 'c6a.2xlarge',
          spotSpec: {
            cpuMax: 8,
            cpuMin: 8,
            memMax: 2,
            memMin: 2,
            disabled: false,
          },
        },
        {
          label: '8x - xlarge - 16VCPU 30GB Memory',
          labelDetail: {
            size: 'xlarge',
            cpu: '16VCPU',
            memory: '32GB',
          },
          value: 'c6a.4xlarge',
          spotSpec: {
            cpuMax: 16,
            cpuMin: 16,
            memMax: 2,
            memMin: 2,
            disabled: false,
          },
        },
      ],
    },
    {
      label: 'General Purpose',
      options: [
        {
          label: '1x - small - 2VCPU 4GB Memory',
          labelDetail: {
            size: 'small',
            cpu: '2VCPU',
            memory: '2GB',
          },
          value: 't4g.small',
          spotSpec: {
            cpuMax: 2,
            cpuMin: 2,
            memMax: 1,
            memMin: 1,
            disabled: false,
          },
        },
        {
          label: '2x - medium - 4VCPU 8GB Memory',
          labelDetail: {
            size: 'medium',
            cpu: '2VCPU',
            memory: '4GB',
          },
          value: 't4g.medium',
          spotSpec: {
            cpuMax: 2,
            cpuMin: 2,
            memMax: 2,
            memMin: 2,
            disabled: false,
          },
        },
        {
          label: '4x - large - 8VCPU 16GB Memory',
          labelDetail: {
            size: 'large',
            cpu: '2VCPU',
            memory: '8GB',
          },
          value: 't4g.large',
          spotSpec: {
            cpuMax: 2,
            cpuMin: 2,
            memMax: 4,
            memMin: 2,
            disabled: false,
          },
        },
        {
          label: '8x - xlarge - 16VCPU 32GB Memory',
          labelDetail: {
            size: 'xlarge',
            cpu: '4VCPU',
            memory: '16GB',
          },
          value: 't4g.xlarge',
          spotSpec: {
            cpuMax: 4,
            cpuMin: 4,
            memMax: 4,
            memMin: 4,
            disabled: false,
          },
        },
      ],
    },
    {
      label: 'GPU Optimised',
      options: [
        {
          label: '1x - small - (1 GPU 24GB) (4VCPU 16GB)',
          labelDetail: {
            size: 'small',
            cpu: '2VCPU',
            memory: '8GB',
          },
          value: 'g5.xlarge',
          spotSpec: {
            cpuMax: 2,
            cpuMin: 2,
            memMax: 4,
            memMin: 4,
            disabled: false,
          },
          gpuEnabled: true,
        },
        {
          label: '2x - medium - (1 GPU 24GB) (8VCPU 32GB)',
          labelDetail: {
            size: 'medium',
            cpu: '4VCPU',
            memory: '16GB',
          },
          value: 'g5.2xlarge',
          spotSpec: {
            cpuMax: 4,
            cpuMin: 4,
            memMax: 4,
            memMin: 4,
            disabled: false,
          },
          gpuEnabled: true,
        },
        {
          label: '4x - large - (1 GPU 24GB) (16VCPU 64GB)',
          labelDetail: {
            size: 'large',
            cpu: '8VCPU',
            memory: '32GB',
          },
          value: 'g5.4xlarge',
          spotSpec: {
            cpuMax: 8,
            cpuMin: 8,
            memMax: 4,
            memMin: 4,
            disabled: false,
          },
          gpuEnabled: true,
        },
        {
          label: '8x - xlarge - (1 GPU 24GB) (32VCPU 128GB)',
          labelDetail: {
            size: 'xlarge',
            cpu: '16VCPU',
            memory: '64GB',
          },
          value: 'g5.8xlarge',
          spotSpec: {
            cpuMax: 16,
            cpuMin: 16,
            memMax: 4,
            memMin: 4,
            disabled: false,
          },
          gpuEnabled: true,
        },
      ],
    },
    {
      label: 'Memory Optimised',
      options: [
        {
          label: '1x - small - 2VCPU 8GB Memory',
          labelDetail: {
            size: 'small',
            cpu: '2VCPU',
            memory: '8GB',
          },
          value: 'm6a.large',
          spotSpec: {
            cpuMax: 2,
            cpuMin: 2,
            memMax: 4,
            memMin: 4,
            disabled: false,
          },
        },
        {
          label: '2x - medium - 4VCPU 16GB Memory',
          labelDetail: {
            size: 'medium',
            cpu: '4VCPU',
            memory: '16GB',
          },
          value: 'm6a.xlarge',
          spotSpec: {
            cpuMax: 4,
            cpuMin: 4,
            memMax: 4,
            memMin: 4,
            disabled: false,
          },
        },
        {
          label: '4x - large - 8VCPU 32GB Memory',
          labelDetail: {
            size: 'large',
            cpu: '8VCPU',
            memory: '32GB',
          },
          value: 'm6a.2xlarge',
          spotSpec: {
            cpuMax: 8,
            cpuMin: 8,
            memMax: 4,
            memMin: 4,
            disabled: false,
          },
        },
        {
          label: '8x - xlarge - 16VCPU 64GB Memory',
          labelDetail: {
            size: 'xlarge',
            cpu: '16VCPU',
            memory: '64GB',
          },
          value: 'm6a.4xlarge',
          spotSpec: {
            cpuMax: 16,
            cpuMin: 16,
            memMax: 4,
            memMin: 4,
            disabled: false,
          },
        },
      ],
    },
  ],
  provisionTypes: [
    { label: 'on-demand', value: 'ec2' },
    { label: 'spot', value: 'spot' },
  ],
  gcpPoolTypes: [
    { label: 'Standard', value: 'STANDARD' },
    { label: 'Spot', value: 'SPOT' },
  ],
  taintsData: [
    { id: 't1', label: 'No execute', value: 'No execute' },
    { id: 't2', label: 'No schedule', value: 'No schedule' },
    {
      id: 't3',
      label: 'Preferred no schedule',
      value: 'Preferred no schedule',
    },
  ],
};

const findNodePlan = (id: string) => {
  return nodePlans
    .flatMap((np) => np.options.flat())
    .find((np) => np.value === id);
};

export const findNodePlanWithSpec = ({
  spot,
  spec,
}: {
  spot?: boolean;
  spec?: {
    cpu?: string;
    memory?: string;
  };
}):
  | ((typeof nodePlans)[number]['options'][number] & { category: string })
  | null => {
  if (!spec) {
    return null;
  }

  let nodePlan = null;

  nodePlans.forEach((np) => {
    np.options.forEach((npp) => {
      if (
        (spot && npp.spotSpec) ||
        (!spot && !npp.spotSpec) ||
        (spot && !npp.spotSpec)
      ) {
        if (
          npp.spotSpec &&
          npp.spotSpec.cpuMax === parseValue(spec.cpu, 0) &&
          npp.spotSpec.memMax === parseValue(spec.memory, 0)
        ) {
          nodePlan = { ...npp, category: np.label };
        }
      }
    });
  });

  return nodePlan;
};

const findNodePlanWithCategory = (
  id: string
):
  | ((typeof nodePlans)[number]['options'][number] & { category: string })
  | null => {
  let nodePlan = null;

  nodePlans.forEach((np) => {
    np.options.forEach((npp) => {
      if (npp.value === id) {
        nodePlan = { ...npp, category: np.label };
      }
    });
  });
  return nodePlan;
};

export {
  nodePlans,
  provisionTypes,
  taintsData,
  findNodePlan,
  findNodePlanWithCategory,
  gcpPoolTypes,
};
