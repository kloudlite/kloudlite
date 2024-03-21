interface Iplan {
  value: string;
  memoryPerCpu: number;
  label: string;
  description: string;
  isShared: boolean;
}

export type IcpuMode = 'shared' | 'dedicated';

export const plans: {
  [key: string]: {
    label: string;
    options: Iplan[];
  };
} = {
  shared: {
    label: 'Shared - Burstable Performance',
    options: [
      {
        isShared: true,
        value: 'shared-1',
        memoryPerCpu: 1,
        label: 'General Purpose',
        description: 'This foundational package for your need',
      },
      {
        isShared: true,
        value: 'shared-2',
        memoryPerCpu: 2,
        label: 'Memory Optmized',
        description: 'This foundational package for your need',
      },
    ],
  },
  dedicated: {
    label: 'Dedicated',
    options: [
      {
        isShared: false,
        value: 'ded-4',
        memoryPerCpu: 4,
        label: 'General Puropse',
        description: 'This foundational package for your need',
      },
      {
        isShared: false,
        value: 'ded-2',
        memoryPerCpu: 2,
        label: 'CPU Optimised',
        description: 'This foundational package for your need',
      },
      {
        isShared: false,
        value: 'ded-8',
        memoryPerCpu: 8,
        label: 'Memory Optimised',
        description: 'This foundational package for your need',
      },
    ],
  },
};
