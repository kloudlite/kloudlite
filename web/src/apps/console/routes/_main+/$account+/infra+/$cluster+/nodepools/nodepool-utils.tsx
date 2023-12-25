const { nodePlans, provisionTypes, taintsData } = {
  nodePlans: [
    {
      label: 'CPU Optimised',
      options: [
        {
          label: '1x - small - 2VCPU 3.75GB Memory',
          labelDetail: {
            size: 'small',
            cpu: '2VCPU',
            memory: '3.75GB',
          },
          value: 'c6a.large',
          spotSpec: {
            cpuMax: 4,
            cpuMin: 4,
            memMax: 8192,
            memMin: 8192,
            disabled: false,
          },
          gpuEnabled: false,
        },
        {
          label: '2x - medium - 4VCPU 7.5GB Memory',
          labelDetail: {
            size: 'medium',
            cpu: '4VCPU',
            memory: '7.5GB',
          },
          value: 'c6a.xlarge',
          spotSpec: {
            cpuMax: 8,
            cpuMin: 8,
            memMax: 16384,
            memMin: 16384,
            disabled: false,
          },
        },
        {
          label: '4x - large - 8VCPU 15GB Memory',
          labelDetail: {
            size: 'large',
            cpu: '8VCPU',
            memory: '15GB',
          },
          value: 'c6a.2xlarge',
          spotSpec: {
            cpuMax: 16,
            cpuMin: 16,
            memMax: 32768,
            memMin: 32768,
            disabled: false,
          },
        },
        {
          label: '8x - xlarge - 16VCPU 30GB Memory',
          labelDetail: {
            size: 'xlarge',
            cpu: '16VCPU',
            memory: '30GB',
          },
          value: 'c6a.4xlarge',
          spotSpec: {
            cpuMax: 32,
            cpuMin: 32,
            memMax: 65536,
            memMin: 65536,
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
            memory: '4GB',
          },
          value: 't4g.small',
          spotSpec: {
            cpuMax: 4,
            cpuMin: 4,
            memMax: 8192,
            memMin: 8192,
            disabled: false,
          },
        },
        {
          label: '2x - medium - 4VCPU 8GB Memory',
          labelDetail: {
            size: 'medium',
            cpu: '4VCPU',
            memory: '8GB',
          },
          value: 't4g.medium',
          spotSpec: {
            cpuMax: 8,
            cpuMin: 8,
            memMax: 16384,
            memMin: 16384,
            disabled: false,
          },
        },
        {
          label: '4x - large - 8VCPU 16GB Memory',
          labelDetail: {
            size: 'large',
            cpu: '8VCPU',
            memory: '16GB',
          },
          value: 't4g.large',
          spotSpec: {
            cpuMax: 16,
            cpuMin: 16,
            memMax: 32768,
            memMin: 32768,
            disabled: false,
          },
        },
        {
          label: '8x - xlarge - 16VCPU 32GB Memory',
          labelDetail: {
            size: 'xlarge',
            cpu: '16VCPU',
            memory: '32GB',
          },
          value: 't4g.xlarge',
          spotSpec: {
            cpuMax: 32,
            cpuMin: 32,
            memMax: 65536,
            memMin: 65536,
            disabled: false,
          },
        },
      ],
    },
    {
      label: 'GPU Optimised',
      options: [
        {
          label: '1x - small - 2VCPU 8GB Memory',
          labelDetail: {
            size: 'small',
            cpu: '2VCPU',
            memory: '8GB',
          },
          value: 'g4dn.xlarge',
          spotSpec: {
            cpuMax: 4,
            cpuMin: 4,
            memMax: 16384,
            memMin: 16384,
            disabled: false,
          },
          gpuEnabled: true,
        },
        {
          label: '2x - medium - 4VCPU 16GB Memory',
          labelDetail: {
            size: 'medium',
            cpu: '4VCPU',
            memory: '16GB',
          },
          value: 'g4dn.2xlarge',
          spotSpec: {
            cpuMax: 8,
            cpuMin: 8,
            memMax: 32768,
            memMin: 32768,
            disabled: false,
          },
          gpuEnabled: true,
        },
        {
          label: '4x - large - 8VCPU 32GB Memory',
          labelDetail: {
            size: 'large',
            cpu: '8VCPU',
            memory: '32GB',
          },
          value: 'g4dn.4xlarge',
          spotSpec: {
            cpuMax: 16,
            cpuMin: 16,
            memMax: 65536,
            memMin: 65536,
            disabled: false,
          },
          gpuEnabled: true,
        },
        {
          label: '8x - xlarge - 16VCPU 64GB Memory',
          labelDetail: {
            size: 'xlarge',
            cpu: '16VCPU',
            memory: '64GB',
          },
          value: 'g4dn.8xlarge',
          spotSpec: {
            cpuMax: 32,
            cpuMin: 32,
            memMax: 131072,
            memMin: 131072,
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
            cpuMax: 4,
            cpuMin: 4,
            memMax: 16384,
            memMin: 16384,
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
            cpuMax: 8,
            cpuMin: 8,
            memMax: 32768,
            memMin: 32768,
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
            cpuMax: 16,
            cpuMin: 16,
            memMax: 65536,
            memMin: 65536,
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
            cpuMax: 32,
            cpuMin: 32,
            memMax: 131072,
            memMin: 131072,
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
};
