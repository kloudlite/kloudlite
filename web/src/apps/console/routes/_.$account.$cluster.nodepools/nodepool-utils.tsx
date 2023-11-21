const { nodePlans, provisionTypes, taintsData } = {
  nodePlans: [
    {
      label: 'CPU Optimised',
      options: [
        {
          label: '1x - small - 2VCPU 3.75GB Memory',
          value: 'c6a.large',
          spotSpec: {
            cpuMax: 4,
            cpuMin: 4,
            memMax: 8192,
            memMin: 8192,
            disabled: false,
          },
        },
      ],
    },
  ],
  provisionTypes: [
    { label: 'On-Demand', value: 'ec2' },
    { label: 'Spot 70% discount', value: 'spot' },
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

export { nodePlans, provisionTypes, taintsData, findNodePlan };
